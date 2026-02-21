package p1tooldisc

import (
	"encoding/binary"
	"fmt"
)

const (
	wtVarint = 0
	wtBytes  = 2
)

type ToolDescriptor struct {
	ToolID            string
	Name              string
	Version           string
	SchemaRef         string
	DescriptorPayload []byte
}

type TooldiscListReq struct {
	PageSize  uint32
	PageToken string
	Filter    string
}

type TooldiscListResp struct {
	Tools         []ToolDescriptor
	NextPageToken string
}

type TooldiscGetReq struct {
	ToolID  string
	Version string
}

type TooldiscGetResp struct {
	Tool ToolDescriptor
}

type TooldiscErr struct {
	Code    string
	Message string
}

func EncodePayloadListReq(v TooldiscListReq) ([]byte, error) {
	return encodeWrapper(1, encodeListReq(v)), nil
}

func DecodePayloadListReq(payload []byte) (TooldiscListReq, error) {
	inner, err := decodeWrapper(payload, 1)
	if err != nil {
		return TooldiscListReq{}, err
	}
	return decodeListReq(inner)
}

func EncodePayloadListResp(v TooldiscListResp) ([]byte, error) {
	return encodeWrapper(2, encodeListResp(v)), nil
}

func DecodePayloadListResp(payload []byte) (TooldiscListResp, error) {
	inner, err := decodeWrapper(payload, 2)
	if err != nil {
		return TooldiscListResp{}, err
	}
	return decodeListResp(inner)
}

func EncodePayloadGetReq(v TooldiscGetReq) ([]byte, error) {
	return encodeWrapper(3, encodeGetReq(v)), nil
}

func DecodePayloadGetReq(payload []byte) (TooldiscGetReq, error) {
	inner, err := decodeWrapper(payload, 3)
	if err != nil {
		return TooldiscGetReq{}, err
	}
	return decodeGetReq(inner)
}

func EncodePayloadGetResp(v TooldiscGetResp) ([]byte, error) {
	return encodeWrapper(4, encodeGetResp(v)), nil
}

func DecodePayloadGetResp(payload []byte) (TooldiscGetResp, error) {
	inner, err := decodeWrapper(payload, 4)
	if err != nil {
		return TooldiscGetResp{}, err
	}
	return decodeGetResp(inner)
}

func EncodePayloadErr(v TooldiscErr) ([]byte, error) {
	return encodeWrapper(5, encodeErr(v)), nil
}

func DecodePayloadErr(payload []byte) (TooldiscErr, error) {
	inner, err := decodeWrapper(payload, 5)
	if err != nil {
		return TooldiscErr{}, err
	}
	return decodeErr(inner)
}

func encodeWrapper(oneofField uint64, inner []byte) []byte {
	var out []byte
	out = appendKey(out, oneofField, wtBytes)
	out = appendBytes(out, inner)
	return out
}

func decodeWrapper(payload []byte, expectedField uint64) ([]byte, error) {
	for len(payload) > 0 {
		field, wt, val, n, err := consumeField(payload)
		if err != nil {
			return nil, err
		}
		if wt == wtBytes && field == expectedField {
			return val, nil
		}
		payload = payload[n:]
	}
	return nil, fmt.Errorf("missing wrapper field %d", expectedField)
}

func encodeListReq(v TooldiscListReq) []byte {
	var out []byte
	if v.PageSize > 0 {
		out = appendKey(out, 1, wtVarint)
		out = binary.AppendUvarint(out, uint64(v.PageSize))
	}
	if v.PageToken != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.PageToken))
	}
	if v.Filter != "" {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, []byte(v.Filter))
	}
	return out
}

func decodeListReq(b []byte) (TooldiscListReq, error) {
	var out TooldiscListReq
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return TooldiscListReq{}, err
		}
		switch field {
		case 1:
			if wt != wtVarint {
				return TooldiscListReq{}, fmt.Errorf("tooldisc_list_req.page_size wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return TooldiscListReq{}, err
			}
			out.PageSize = uint32(vv)
		case 2:
			if wt != wtBytes {
				return TooldiscListReq{}, fmt.Errorf("tooldisc_list_req.page_token wrong wire type")
			}
			out.PageToken = string(val)
		case 3:
			if wt != wtBytes {
				return TooldiscListReq{}, fmt.Errorf("tooldisc_list_req.filter wrong wire type")
			}
			out.Filter = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeListResp(v TooldiscListResp) []byte {
	var out []byte
	for _, td := range v.Tools {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, encodeToolDescriptor(td))
	}
	if v.NextPageToken != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.NextPageToken))
	}
	return out
}

func decodeListResp(b []byte) (TooldiscListResp, error) {
	var out TooldiscListResp
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return TooldiscListResp{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return TooldiscListResp{}, fmt.Errorf("tooldisc_list_resp.tools wrong wire type")
			}
			td, err := decodeToolDescriptor(val)
			if err != nil {
				return TooldiscListResp{}, err
			}
			out.Tools = append(out.Tools, td)
		case 2:
			if wt != wtBytes {
				return TooldiscListResp{}, fmt.Errorf("tooldisc_list_resp.next_page_token wrong wire type")
			}
			out.NextPageToken = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeGetReq(v TooldiscGetReq) []byte {
	var out []byte
	if v.ToolID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.ToolID))
	}
	if v.Version != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Version))
	}
	return out
}

func decodeGetReq(b []byte) (TooldiscGetReq, error) {
	var out TooldiscGetReq
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return TooldiscGetReq{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return TooldiscGetReq{}, fmt.Errorf("tooldisc_get_req.tool_id wrong wire type")
			}
			out.ToolID = string(val)
		case 2:
			if wt != wtBytes {
				return TooldiscGetReq{}, fmt.Errorf("tooldisc_get_req.version wrong wire type")
			}
			out.Version = string(val)
		}
		b = b[n:]
	}
	if out.ToolID == "" {
		return TooldiscGetReq{}, fmt.Errorf("tooldisc_get_req.tool_id required")
	}
	return out, nil
}

func encodeGetResp(v TooldiscGetResp) []byte {
	var out []byte
	out = appendKey(out, 1, wtBytes)
	out = appendBytes(out, encodeToolDescriptor(v.Tool))
	return out
}

func decodeGetResp(b []byte) (TooldiscGetResp, error) {
	var out TooldiscGetResp
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return TooldiscGetResp{}, err
		}
		if field == 1 {
			if wt != wtBytes {
				return TooldiscGetResp{}, fmt.Errorf("tooldisc_get_resp.tool wrong wire type")
			}
			td, err := decodeToolDescriptor(val)
			if err != nil {
				return TooldiscGetResp{}, err
			}
			out.Tool = td
		}
		b = b[n:]
	}
	return out, nil
}

func encodeErr(v TooldiscErr) []byte {
	var out []byte
	if v.Code != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.Code))
	}
	if v.Message != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Message))
	}
	return out
}

func decodeErr(b []byte) (TooldiscErr, error) {
	var out TooldiscErr
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return TooldiscErr{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return TooldiscErr{}, fmt.Errorf("tooldisc_err.code wrong wire type")
			}
			out.Code = string(val)
		case 2:
			if wt != wtBytes {
				return TooldiscErr{}, fmt.Errorf("tooldisc_err.message wrong wire type")
			}
			out.Message = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeToolDescriptor(v ToolDescriptor) []byte {
	var out []byte
	if v.ToolID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.ToolID))
	}
	if v.Name != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Name))
	}
	if v.Version != "" {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, []byte(v.Version))
	}
	if v.SchemaRef != "" {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, []byte(v.SchemaRef))
	}
	if len(v.DescriptorPayload) > 0 {
		out = appendKey(out, 5, wtBytes)
		out = appendBytes(out, v.DescriptorPayload)
	}
	return out
}

func decodeToolDescriptor(b []byte) (ToolDescriptor, error) {
	var out ToolDescriptor
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return ToolDescriptor{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return ToolDescriptor{}, fmt.Errorf("tool_descriptor.tool_id wrong wire type")
			}
			out.ToolID = string(val)
		case 2:
			if wt != wtBytes {
				return ToolDescriptor{}, fmt.Errorf("tool_descriptor.name wrong wire type")
			}
			out.Name = string(val)
		case 3:
			if wt != wtBytes {
				return ToolDescriptor{}, fmt.Errorf("tool_descriptor.version wrong wire type")
			}
			out.Version = string(val)
		case 4:
			if wt != wtBytes {
				return ToolDescriptor{}, fmt.Errorf("tool_descriptor.schema_ref wrong wire type")
			}
			out.SchemaRef = string(val)
		case 5:
			if wt != wtBytes {
				return ToolDescriptor{}, fmt.Errorf("tool_descriptor.descriptor_payload wrong wire type")
			}
			out.DescriptorPayload = append([]byte(nil), val...)
		}
		b = b[n:]
	}
	return out, nil
}

func appendKey(dst []byte, field, wt uint64) []byte {
	return binary.AppendUvarint(dst, (field<<3)|wt)
}

func appendBytes(dst []byte, b []byte) []byte {
	dst = binary.AppendUvarint(dst, uint64(len(b)))
	return append(dst, b...)
}

func consumeField(b []byte) (field uint64, wt uint64, val []byte, consumed int, err error) {
	key, n := binary.Uvarint(b)
	if n <= 0 {
		return 0, 0, nil, 0, fmt.Errorf("invalid protobuf key")
	}
	field = key >> 3
	wt = key & 0x7
	idx := n

	switch wt {
	case wtVarint:
		_, vn := binary.Uvarint(b[idx:])
		if vn <= 0 {
			return 0, 0, nil, 0, fmt.Errorf("invalid protobuf varint value")
		}
		return field, wt, b[idx : idx+vn], idx + vn, nil
	case wtBytes:
		l, ln := binary.Uvarint(b[idx:])
		if ln <= 0 {
			return 0, 0, nil, 0, fmt.Errorf("invalid protobuf bytes length")
		}
		idx += ln
		if idx+int(l) > len(b) {
			return 0, 0, nil, 0, fmt.Errorf("truncated protobuf bytes value")
		}
		return field, wt, b[idx : idx+int(l)], idx + int(l), nil
	default:
		return 0, 0, nil, 0, fmt.Errorf("unsupported protobuf wire type %d", wt)
	}
}

func consumeVarintValue(b []byte) (uint64, int, error) {
	v, n := binary.Uvarint(b)
	if n <= 0 {
		return 0, 0, fmt.Errorf("invalid varint value")
	}
	return v, n, nil
}
