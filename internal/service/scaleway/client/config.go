package client

type Config interface {
	GetSecretKey() string
}

func (c *Client) GetSecretKey() string {
	return c.secretKey
}
