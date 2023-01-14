package swatch

import (
	"encoding/json"
	"net"

	"github.com/pkg/errors"
)

type Client struct {
	Conn net.Conn
}

func NewClient() (Client, error) {
	c, err := net.Dial("unix", SwatchSocketPath)
	if err != nil {
		return Client{}, errors.Wrapf(err, "could not connect to socket %s", SwatchSocketPath)
	}
	return Client{
		Conn: c,
	}, nil
}

func (c *Client) Write(b []byte) error {
	_, err := c.Conn.Write(b)
	if err != nil {
		return errors.Wrap(err, "could not write bytes")
	}
	return nil
}

func (c *Client) Read() (Message, error) {
	buf := make([]byte, 1024)
	n, err := c.Conn.Read(buf)
	if err != nil {
		return Message{}, errors.Wrapf(err, "could not read from socket %s", SwatchSocketPath)
	}
	message := Message{}
	err = json.Unmarshal(buf[:n], &message)
	if err != nil {
		return Message{}, errors.Wrap(err, "could not unmarshal bytes to message struct")
	}
	return message, nil
}
