package main

func main() {
	srv := KalibratorGoServer(8080) 
	if err := srv.ListenAndServe(); err != nil {
		panic("server failed to start: " + err.Error())
	}
}
