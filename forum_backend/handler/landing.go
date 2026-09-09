package handler

import (
	"net/http"
)

func GetLanding(w http.ResponseWriter, r *http.Request) {
	jsonMock := `[
		{
	        "id": 1,
	        "title": "Hello world",
	        "content": "My first post"
	    },
	    {
	        "id": 2,
	        "title": "Second post",
	        "content": "Another post"
	    }
    ]`

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonMock))
}
