package generator

import (
	"fmt"
	"math/rand"

	gofakeit "github.com/brianvoe/gofakeit/v4"
)

func payload(httpMethod string, responseCode int, uri string) string {
	return fmt.Sprintf("[%d] %s %s", responseCode, httpMethod, uri)
}

func Generate(records int64) []string {
	var values []string
	codes := []int{200, 301, 403, 404, 500}
	i := int64(0)
	for {
		pickCode := codes[rand.Intn(len(codes))]
		payload := payload(gofakeit.HTTPMethod(), pickCode, fmt.Sprintf("/%s/%s", gofakeit.BuzzWord(), gofakeit.BuzzWord()))
		values = append(values, payload)
		i++
		if records == i {
			break
		}
	}
	return values
}
