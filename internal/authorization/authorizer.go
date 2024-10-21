package authorization

type Authorizer interface {
	ProduceToken(uid string) (string, error)
	VerifyToken(ts string) (string, error)
}
