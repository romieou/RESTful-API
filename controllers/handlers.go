package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"rest/models"
	"rest/myerrors"
	"rest/utils"
	"rest/viewmodels"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/semaphore"
)

type MyServer struct {
	db        models.MySQLInterface
	redisConn models.RedisInterface
	jobQueue  chan job
	workers   *workers
}

type job struct {
	ID   string
	hash int64
}

type workers struct {
	mx  *sync.Mutex
	sem *semaphore.Weighted
}

// NewMyServer returns MyServer instance for given MySQK and RedisCache
func NewMyServer(db models.MySQLInterface, r models.RedisInterface) *MyServer {
	return &MyServer{
		db:        db,
		redisConn: r,
		jobQueue:  make(chan job, 2048),
		workers: &workers{
			mx:  &sync.Mutex{},
			sem: semaphore.NewWeighted(2),
		},
	}
}

const (
	dir        = "./"
	emailMsg   = "To parse emails, follow the /check endpoint."
	hashMsg    = "Send a plain string as body of POST request to /rest/hash/calc where you will receive a unique ID.\nUse that ID to get hash with GET request from /rest/hash/result/$id"
	N          = 10
	pendingMsg = "PENDING"
	substrMsg  = "To get the longest substring, follow the /find endpoint."
	successMsg = "Success!"
)

// SubstringHandler handles /rest/substr path
func (s *MyServer) SubstringHandler(ctx *fasthttp.RequestCtx) {
	viewmodels.Message(ctx, substrMsg)
}

// GetSubstring handles the /rest/substr/find path returning the longest substringwith unique chracters
func (s *MyServer) GetSubstring(ctx *fasthttp.RequestCtx) {
	bodyBytes := ctx.Request.Body()
	var str string
	if err := json.Unmarshal(bodyBytes, &str); err != nil {
		log.Println("GetSubstring err:", err)
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	log.Printf("GetSubstring: received string %q", str)
	if str == "" || !utils.IsLatin(str) {
		log.Println("Invalid or empty string input:", str)
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	substr := utils.LongestSubstring(str)
	viewmodels.Message(ctx, substr)
}

// EmailHandler handles /rest/email path
func (s *MyServer) EmailHandler(ctx *fasthttp.RequestCtx) {
	viewmodels.Message(ctx, emailMsg)
}

// GetEmail parses string input and outputs all valid emails separated by comma
// Acceptable format is "Email:_/n/remail@gmail.com"
func (s *MyServer) GetEmail(ctx *fasthttp.RequestCtx) {
	var email string
	bodyBytes := ctx.Request.Body()
	if err := json.Unmarshal(bodyBytes, &email); err != nil {
		log.Println("GetEmail err:", err)
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	log.Println("received string:", email)
	re := regexp.MustCompile(`Email:[_\r\n]+(?P<email>[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4})`)

	matches := re.FindAll([]byte(email), -1)
	if len(matches) == 0 {
		log.Println("GetEmail: match not found")
		viewmodels.ClientError(ctx, fasthttp.StatusNotFound, myerrors.ErrInvalidInput)
		return
	}

	// res stores comma-separated emails
	var res string
	for i, v := range matches {
		res += string(re.ReplaceAll(v, []byte("$email")))
		if i != len(matches)-1 {
			res += ", "
		}
	}
	viewmodels.Message(ctx, res)
}

// GetIIN parses string input and outputs all valid IINs separated by space
// Acceptable format is "IIN:_/n/rvalidIIN"
func (s *MyServer) GetIIN(ctx *fasthttp.RequestCtx) {
	var IIN string
	bodyBytes := ctx.Request.Body()
	if err := json.Unmarshal(bodyBytes, &IIN); err != nil {
		log.Println("GetIIN err:", err)
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}

	log.Println("received IIN:", IIN)
	// IIN cannot be followed by digit(s)
	re := regexp.MustCompile(`IIN:[_\r\n]+(?P<iin>\d{12})([\D]|\z)`)

	matches := re.FindAll([]byte(IIN), -1)
	if len(matches) == 0 {
		log.Println("GetIIN: match not found")
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}

	// res stores space-separated IINs
	var res string
	for i, v := range matches {
		curr := string(re.ReplaceAll(v, []byte("$iin")))
		if utils.ValidateIIN(curr) {
			res += curr
			if i != len(matches)-1 {
				res += " "
			}
		}

	}
	viewmodels.Message(ctx, res)
}

// Add implements addition to counter.
// The function accepts numbers with leading zeroes and negative numbers.
func (s *MyServer) Add(ctx *fasthttp.RequestCtx, n int) {
	res, err := s.redisConn.SetCounter(n)
	if err != nil {
		log.Println("Add err:", err)
		if err == myerrors.ErrNegativeCounter {
			viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, err)
			return
		}
		viewmodels.ServerError(ctx)
		return
	}
	viewmodels.Message(ctx, successMsg+" Counter is now "+res)
}

// AddCounter adds the number in path to counter
func (s *MyServer) AddCounter(ctx *fasthttp.RequestCtx) {
	addVal, ok := ctx.UserValue("add").(string)
	if !ok {
		log.Println("Couldn't get add value from context")
		viewmodels.ServerError(ctx) // maybe wrap around more context
		return
	}
	n, err := strconv.Atoi(addVal)
	if err != nil {
		log.Println("Invalid add value:", addVal)
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	s.Add(ctx, n)

}

// SubCounter substract the number in path from counter
func (s *MyServer) SubCounter(ctx *fasthttp.RequestCtx) {
	subVal, ok := ctx.UserValue("sub").(string)
	if !ok {
		log.Println("Couldn't get sub value from context")
		viewmodels.ServerError(ctx) // maybe wrap around more context
		return
	}
	n, err := strconv.Atoi(subVal)
	if err != nil {
		log.Println("Invalid subVal:", subVal)
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	// check for overflow
	if n < 0 && n*-1 < 0 || n > 0 && n*-1 > 0 {
		log.Println("Provided sub value too large")
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("%w, number is too large", myerrors.ErrInvalidInput))
		return
	}
	n *= -1
	s.Add(ctx, n)
}

// GetCounter gets counter's current value
func (s *MyServer) GetCounter(ctx *fasthttp.RequestCtx) {
	counter, err := s.redisConn.GetCounter()
	if err != nil {
		log.Println("GetCounter err:", err)
		viewmodels.ServerError(ctx)
		return
	}
	viewmodels.Message(ctx, fmt.Sprintf("counter value is %s", counter))
}

// CreateUser creates new user for provided first- and lastname
// Request body should be structured as JSON with "first_name" and "last_name"
// Body must contain both the first- and lastname
func (s *MyServer) CreateUser(ctx *fasthttp.RequestCtx) {
	var user models.User
	bodyBytes := ctx.Request.Body()
	if len(bodyBytes) == 0 {
		log.Println("Couldn't get body")
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrBodyNotFound)
		return
	}
	if err := json.Unmarshal(bodyBytes, &user); err != nil || !utils.ValidateUser(user) {
		log.Println("Invalid user input:", string(bodyBytes))
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("%w Provide first_name and last_name", myerrors.ErrInvalidInput))
		return
	}
	id, err := s.db.CreateUser(&user)
	if err != nil {
		log.Println("Failed to create user:", err)
		viewmodels.ServerError(ctx)
		return
	}
	viewmodels.Message(ctx, fmt.Sprintf("%s Created new user under ID %d", successMsg, id))

}

// GetUser retrieves information on user under the provided ID
func (s *MyServer) GetUser(ctx *fasthttp.RequestCtx) {
	ID, ok := ctx.UserValue("id").(string)
	if !ok {
		log.Println("Couldn't get ID from context")
		viewmodels.ServerError(ctx)
		return
	}
	if !utils.ValidateID(ID) {
		log.Println("Invalid ID:", ID)
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	user, err := s.db.GetUser(ID)
	if err != nil {
		log.Println("GetUser err:", err)
		if err == myerrors.ErrUserNotFound {
			viewmodels.ClientError(ctx, fasthttp.StatusNotFound, err)
			return
		}
		viewmodels.ServerError(ctx)
		return
	}
	viewmodels.JSON(ctx, user)
}

// UpdateUser updates user by ID to provided data
// Request body should be structured as JSON with "first_name" and "last_name"
// Body should contain either first- or lastname, both cannot be empty
func (s *MyServer) UpdateUser(ctx *fasthttp.RequestCtx) {
	ID, ok := ctx.UserValue("id").(string)
	if !ok {
		log.Println("Couldn't get ID from context")
		viewmodels.ServerError(ctx)
		return
	}
	if !utils.ValidateID(ID) {
		log.Println("Invalid ID:", ID)
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	bodyBytes := ctx.Request.Body()
	if len(bodyBytes) == 0 {
		log.Println("Couldn't get body")
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	var user models.User
	if err := json.Unmarshal(bodyBytes, &user); err != nil {
		log.Println("Invalid user input:", string(bodyBytes))
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	if firstName, lastName := user.FirstName, user.LastName; firstName != "" && !utils.IsLatin(firstName) || lastName != "" && !utils.IsLatin(lastName) || firstName == "" && lastName == "" {
		log.Println("Invalid first- or lastname or both are empty")
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	if err := s.db.UpdateUser(ID, user); err != nil {
		log.Println(err)
		viewmodels.ServerError(ctx)
		return
	}
	viewmodels.Message(ctx, fmt.Sprintf("%s Updated user under ID %s. To view changes, go to /rest/user/%s.", successMsg, ID, ID))
}

// DeleteUser deletes user, if such exists, by ID
func (s *MyServer) DeleteUser(ctx *fasthttp.RequestCtx) {
	ID, ok := ctx.UserValue("id").(string)
	if !ok {
		log.Println("Couldn't get ID from context")
		viewmodels.ServerError(ctx)
		return
	}
	if !utils.ValidateID(ID) {
		log.Println("User provided invalid ID")
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	if err := s.db.DeleteUser(ID); err != nil {
		log.Println("DeleteUser err:", err)
		if err == myerrors.ErrUserNotFound {
			viewmodels.ClientError(ctx, fasthttp.StatusNotFound, err)
			return
		}
		viewmodels.ServerError(ctx)
		return
	}
	viewmodels.Message(ctx, fmt.Sprintf("%s Deleted user under ID %s", successMsg, ID))
}

// HashHandler hadnles /rest/hash
func (s *MyServer) HashHandler(ctx *fasthttp.RequestCtx) {
	viewmodels.Message(ctx, hashMsg)
}

// GenerateHash handles /rest/hash/calc
func (s *MyServer) GenerateHash(ctx *fasthttp.RequestCtx) {

	var strInput string
	bodyBytes := ctx.Request.Body()
	if len(bodyBytes) == 0 {
		log.Println("Generate hash err: empty body")
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	if err := json.Unmarshal(bodyBytes, &strInput); err != nil || strInput == "" {
		log.Println("Generate hash err:", err)
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	hash, err := strconv.ParseInt(strInput, 10, 64)
	if err != nil {
		log.Println("Generate hash err:", err)
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	ID := uuid.New().String()
	//viewmodels.Message(ctx, fmt.Sprintf("Your id is %s", ID))
	log.Println("generated uuid", ID)
	s.redisConn.Set(ID, pendingMsg)
	s.jobQueue <- job{ID, hash}
	viewmodels.Message(ctx, fmt.Sprintf("We have received your request and assigned the ID %s", ID))
}

// MakeHash implements hash generation logic
func (s *MyServer) MakeHash(ctx context.Context, hash int64, ID string) error {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	defer s.workers.sem.Release(1)
	for {
		select {
		case <-ticker.C:
			fmt.Println(ID)
			nsec := s.workers.GetTimestamp()
			hash = hash & nsec
		case <-ctx.Done():
			res := strconv.Itoa(utils.CountBits(hash))
			log.Println(fmt.Sprintf("Generated hash %s for ID %s", res, ID))
			return s.redisConn.Set(ID, res)
		}
	}
}

// GetTimestamp gets current timestamp
func (w *workers) GetTimestamp() int64 {
	w.mx.Lock()
	defer w.mx.Unlock()
	now := time.Now()
	return int64(now.UnixNano())
}

// GetHash retrieves hash for given ID
// If hash is not yet generated, it returns "PENDING"
func (s *MyServer) GetHash(ctx *fasthttp.RequestCtx) {
	ID, ok := ctx.UserValue("id").(string)
	if !ok {
		log.Println("GetHash: couldn't get ID value from context")
		viewmodels.ServerError(ctx)
		return
	}
	if ID == "" {
		log.Println("GetHash: invalid ID")
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	hash, err := s.redisConn.Get(ID)
	if err != nil {
		if err == myerrors.ErrNotFound {
			log.Println("GetHash: ID doesn't exist")
			viewmodels.ClientError(ctx, fasthttp.StatusNotFound, myerrors.ErrInvalidInput)
			return
		}
		log.Println("GetHash err:", err)
		viewmodels.ServerError(ctx)
		return
	}
	viewmodels.Message(ctx, fmt.Sprintf("Your hash is %s", hash))
}

//DOCKER_BUILDKIT=1 docker build .

// DispatchWorkers runs workers upon server initialization waiting for tasks
func (s *MyServer) DispatchWorkers() {
	for job := range s.jobQueue {
		// must defer cancel!
		c, _ := context.WithTimeout(context.Background(), time.Minute)
		if err := s.workers.sem.Acquire(c, 1); err != nil {
			fmt.Println(fmt.Errorf("wait for resources: %w", err))
		}
		go s.MakeHash(c, job.hash, job.ID)
	}
}

// GetIdentifiers finds all identifiers with specified name
func (s *MyServer) GetIdentifiers(ctx *fasthttp.RequestCtx) {
	str, ok := ctx.UserValue("str").(string)
	if !ok {
		log.Println("GetIdentifiers: couldn't get ID value from context")
		viewmodels.ServerError(ctx)
		return
	}
	if str == "" {
		log.Println("GetIdentifiers: invalid ID")
		viewmodels.ClientError(ctx, fasthttp.StatusBadRequest, myerrors.ErrInvalidInput)
		return
	}
	fmt.Println(str, dir)
	res, err := utils.GetIdentifiers(str, dir)
	if err != nil {
		log.Println("GetIdentifiers err:", err)
		viewmodels.ServerError(ctx)
		return
	}
	viewmodels.Message(ctx, string(res))

}
