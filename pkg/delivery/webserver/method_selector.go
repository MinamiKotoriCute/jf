package webserver

func MethodSelectorGet(method string) bool {
	return method == "GET"
}

func MethodSelectorPost(method string) bool {
	return method == "POST"
}

func MethodSelectorGetPost(method string) bool {
	return method == "GET" || method == "POST"
}
