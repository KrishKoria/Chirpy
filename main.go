package main

import "net/http"

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8");
	w.WriteHeader(http.StatusOK);
	w.Write([]byte("OK"));
}

func main()  {
	mux := http.NewServeMux();
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("./app"))));
	mux.HandleFunc("/healthz", readinessHandler);
	
	server:= &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	server.ListenAndServe();
}