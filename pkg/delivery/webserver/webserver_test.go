package webserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/MinamiKotoriCute/jf/internal/pb"
)

func loginREQ(ctx context.Context, req *pb.LoginREQ) (*pb.LoginRSP, error) {
	return &pb.LoginRSP{
		ErrorMessage: "Login fail",
	}, nil
}

func login2REQ(ctx context.Context, req *pb.Login2REQ) (*pb.Login2RSP, error) {
	return &pb.Login2RSP{
		ErrorMessage: "Login2 fail",
	}, nil
}

func TestWebServer(t *testing.T) {
	o := NewWebServer()

	o.HandleGetFuncs("/web",
		loginREQ,
		login2REQ)

	if err := o.Start(":8080"); err != nil {
		t.Fatal(err)
	}
	defer o.Stop(context.Background())

	time.Sleep(time.Second * 1)

	r, err := http.Get("http://127.0.0.1:8080/web/pb.LoginREQ")
	if err != nil {
		t.Fatal(err)
		return
	}

	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
		return
	}

	if string(data) != `{"error_message":"Login fail"}` {
		t.Fatal("not match")
	}

	r, err = http.Get("http://127.0.0.1:8080/web/pb.Login2REQ")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer r.Body.Close()
	data, err = io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	if string(data) != `{"error_message":"Login2 fail"}` {
		t.Fatal("not match")
	}
}
