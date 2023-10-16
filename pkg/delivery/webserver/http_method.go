package webserver

type HttpMethod int

const (
	HttpMethodGet HttpMethod = 1 << iota
	MHttpMethodPost
	HttpMethodOptions
)

func MethodStringToHttpMethod(method string) HttpMethod {
	switch method {
	case "GET":
		return HttpMethodGet
	case "POST":
		return MHttpMethodPost
	case "OPTIONS":
		return HttpMethodOptions
	default:
		return 0
	}
}

func HttpMethodToMethodString(httpMethod HttpMethod) string {
	switch httpMethod {
	case HttpMethodGet:
		return "GET"
	case MHttpMethodPost:
		return "POST"
	case HttpMethodOptions:
		return "OPTIONS"
	default:
		return ""
	}
}

func (o HttpMethod) Match(method string) bool {
	return o&MethodStringToHttpMethod(method) != 0
}

func (o HttpMethod) ToString() string {
	ret := ""
	for i := 0; i < 3; i++ {
		httpMethod := HttpMethod(1 << i)
		if o&httpMethod != 0 {
			ret += HttpMethodToMethodString(httpMethod) + ", "
		}
	}

	return ret[:len(ret)-2]
}
