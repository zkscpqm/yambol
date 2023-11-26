package httpx

type Response interface {
	GetStatusCode() int
	Render() ([]byte, error)
}
