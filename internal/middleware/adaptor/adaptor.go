package adaptor

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func GenericDecoder[T any](r *http.Request) (in T, err error) {
	ptrIn := new(T)
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(ptrIn)
	in = *ptrIn
	return in, err
}

func GenericEncoder[T any](w http.ResponseWriter, out T) error {
	json.NewEncoder(w).Encode(out)
	w.Header().Add("Content-Type", "application/json")
	return nil
}

func GenericHttpAdaptor[TIN any, TOUT any](f func(context.Context, TIN) (TOUT, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		in, err := GenericDecoder[TIN](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			zap.L().Sugar().Error(err, r)
			return
		}
		out, err := f(r.Context(), in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			zap.L().Sugar().Error(err, r)
			return
		}
		err = GenericEncoder(w, out)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			zap.L().Sugar().Error(err, r)
		}
	}
}
