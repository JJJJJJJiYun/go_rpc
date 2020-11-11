package codec

import (
	"bufio"
	"encoding/json"
	"io"
)

type JsonCodec struct {
	conn    io.ReadWriteCloser
	buf     *bufio.Writer
	decoder *json.Decoder
	encoder *json.Encoder
}

func (j *JsonCodec) Close() error {
	return j.conn.Close()
}

func (j *JsonCodec) ReadHeader(header *Header) error {
	return j.decoder.Decode(header)
}

func (j *JsonCodec) ReadBody(body interface{}) error {
	return j.decoder.Decode(body)
}

func (j *JsonCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		_ = j.buf.Flush()
		if err != nil {
			_ = j.Close()
		}
	}()
	if err := j.encoder.Encode(header); err != nil {
		return err
	}
	if err := j.encoder.Encode(body); err != nil {
		return err
	}
	return nil
}

func NewJsonCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &JsonCodec{
		conn:    conn,
		buf:     buf,
		decoder: json.NewDecoder(conn),
		encoder: json.NewEncoder(buf),
	}
}
