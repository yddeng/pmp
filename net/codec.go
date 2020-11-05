package net

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet/drpc"
	"github.com/yddeng/dutil/buffer"
	"github.com/yddeng/pmp/net/pb"
	_ "github.com/yddeng/pmp/protocol"
	"io"
	"reflect"
)

var ErrTooLarge = fmt.Errorf("Message too large")

const (
	flagSize = 1
	seqSize  = 8
	cmdSize  = 2
	bodySize = 4
	HeadSize = flagSize + seqSize + cmdSize + bodySize
	BuffSize = 1024 * 256
)

type flag byte

const (
	message flag = 0x80
	rpcReq  flag = 0x40
	rpcResp flag = 0x20
)

func (this flag) setType(tt flag) flag {
	switch tt {
	case message:
		return this | message
	case rpcReq:
		return this | rpcReq
	case rpcResp:
		return this | rpcResp
	default:
		panic("invalid type")
	}
}

func (this flag) getType() flag {
	if this&message > 0 {
		return message
	} else if this&rpcReq > 0 {
		return rpcReq
	} else if this&rpcResp > 0 {
		return rpcResp
	}
	return message
}

// 低第二位
func (this flag) setFile() flag {
	return this | 0x2
}

func (this flag) isFile() bool {
	return this&0x2 > 0
}

// 低第一位
func (this flag) setNext() flag {
	return this | 0x1
}

func (this flag) hasNext() bool {
	return this&0x1 > 0
}

type Codec struct {
	ss, req, resp string
	readBuf       *buffer.Buffer
	readHead      bool
	flag          byte
	seqNo         uint64
	cmd           uint16
	bodyLen       uint32
}

func NewCodec(ss, req, resp string) *Codec {
	return &Codec{
		ss:      ss,
		req:     req,
		resp:    resp,
		readBuf: buffer.NewBufferWithCap(BuffSize),
	}
}

//解码
func (decoder *Codec) Decode(reader io.Reader) (interface{}, error) {
	for {
		msg, err := decoder.unPack()

		if msg != nil {
			return msg, nil

		} else if err == nil {
			_, err1 := decoder.readBuf.ReadFrom(reader)
			if err1 != nil {
				return nil, err1
			}
		} else {
			return nil, err
		}
	}
}

func (decoder *Codec) unPack() (interface{}, error) {

	if !decoder.readHead {
		if decoder.readBuf.Len() < HeadSize {
			return nil, nil
		}

		decoder.flag, _ = decoder.readBuf.ReadByte()
		decoder.seqNo, _ = decoder.readBuf.ReadUint64BE()
		decoder.cmd, _ = decoder.readBuf.ReadUint16BE()
		decoder.bodyLen, _ = decoder.readBuf.ReadUint32BE()
		decoder.readHead = true
	}

	if decoder.bodyLen > BuffSize-HeadSize {
		return nil, ErrTooLarge
	}
	if decoder.readBuf.Len() < int(decoder.bodyLen) {
		return nil, nil
	}

	data, _ := decoder.readBuf.ReadBytes(int(decoder.bodyLen))

	var msg interface{}

	tt := flag(decoder.flag).getType()
	switch tt {
	case message:
		m, err := pb.Unmarshal(decoder.ss, decoder.cmd, data)
		if err != nil {
			return nil, err
		}
		msg = &Message{
			data: m.(proto.Message),
			cmd:  decoder.cmd,
			Next: flag(decoder.flag).hasNext(),
		}
	case rpcReq:
		m, err := pb.Unmarshal(decoder.req, decoder.cmd, data)
		if err != nil {
			return nil, err
		}
		msg = &drpc.Request{
			SeqNo:    decoder.seqNo,
			Method:   pb.GetNameById(decoder.req, decoder.cmd),
			Data:     m,
			NeedResp: true,
		}
	case rpcResp:
		m, err := pb.Unmarshal(decoder.resp, decoder.cmd, data)
		if err != nil {
			return nil, err
		}
		msg = &drpc.Response{
			SeqNo: decoder.seqNo,
			Data:  m,
		}
	}

	decoder.readHead = false
	return msg, nil
}

//编码
func (encoder *Codec) Encode(o interface{}) ([]byte, error) {
	var tt flag
	var seqNo uint64
	var cmd uint16
	var data []byte
	var bodyLen int
	var err error

	switch o.(type) {
	case *Message:
		msg := o.(*Message)
		tt = tt.setType(message)
		cmd, data, err = pb.Marshal(encoder.ss, msg.GetData())
		if err != nil {
			return nil, err
		}
		if msg.Next {
			tt = tt.setNext()
		}
	case *drpc.Request:
		msg := o.(*drpc.Request)
		tt = tt.setType(rpcReq)
		seqNo = msg.SeqNo
		cmd, data, err = pb.Marshal(encoder.req, msg.Data)
		if err != nil {
			return nil, err
		}

	case *drpc.Response:
		msg := o.(*drpc.Response)
		if msg.Err == nil {
			tt = tt.setType(rpcResp)
			seqNo = msg.SeqNo
			cmd, data, err = pb.Marshal(encoder.resp, msg.Data)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("invalid type:%s", reflect.TypeOf(o).String())
	}

	bodyLen = len(data)
	if bodyLen > BuffSize-HeadSize {
		return nil, ErrTooLarge
	}

	totalLen := HeadSize + bodyLen
	buff := buffer.NewBufferWithCap(totalLen)
	//tt
	buff.WriteUint8BE(byte(tt))
	//seq
	buff.WriteUint64BE(seqNo)
	//cmd
	buff.WriteUint16BE(cmd)
	//bodylen
	buff.WriteUint32BE(uint32(bodyLen))
	//body
	buff.WriteBytes(data)

	return buff.Bytes(), nil
}
