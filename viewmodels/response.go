package viewmodels

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

const serverErrorMsg = "Something went wrong. Please try again later."

func Message(ctx *fasthttp.RequestCtx, message string) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.WriteString(message)
}

func ClientError(ctx *fasthttp.RequestCtx, status int, err error) {
	ctx.SetStatusCode(status)
	ctx.WriteString(err.Error())
}

func ServerError(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.WriteString(serverErrorMsg)
}

func JSON(ctx *fasthttp.RequestCtx, data interface{}) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	json.NewEncoder(ctx).Encode(data)
}
