package codec

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
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
		if flushErr := j.buf.Flush(); flushErr != nil {
			log.Println("JsonCodec Write flush buf err: ", err)
			err = flushErr
		}
		if err != nil {
			_ = j.Close()
		}
	}()
	if err := j.encoder.Encode(header); err != nil {
		log.Printf("JsonCodec Write encode header failed. header: %v, err: %v", header, err)
		return err
	}
	if err := j.encoder.Encode(body); err != nil {
		log.Printf("JsonCodec Write encode body failed. header: %v, err: %v", header, err)
		return err
	}
	return nil
}

func newJsonCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &JsonCodec{
		conn:    conn,
		buf:     buf,
		decoder: json.NewDecoder(conn),
		encoder: json.NewEncoder(buf),
	}
}
