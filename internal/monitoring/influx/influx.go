package influx

import (
	"strconv"
	"time"
)

type tag struct {
	Key   string
	Value string
}

type value struct{
	Key   string
	Value interface{}
}

type point struct {
	table     string
	tags      []*Tag
	values    []*Value
	timestamp time.Time
}

type influxClient {
	address string
	bufferCh chan string //Point ?
	writeCh      chan string
	writeBuffer []string

	// writeStop    chan int
	// bufferStop   chan int
	// bufferFlush  chan int
	// doneCh       chan int
	// errCh        chan error
	// bufferInfoCh chan writeBuffInfoReq
	// writeInfoCh  chan writeBuffInfoReq
}

func NewInfluxClient(address string) *MonClient {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}
	client:=&Client{
		
		address:address,
		bufferCh:make(chan Point)
	}
	//go client.bufferProc()
	return client
}

func (c *Client) WritePoint(point *Point) {
	//w.bufferCh <- point.ToLineProtocol(w.service.clientImpl.Options().Precision)
	line, err := point.String()
	if err != nil {
		logger.Errorf("point encoding error: %s\n", err.Error())
	} else {
		c.bufferCh <- line
	}
}

func (w *writeApiImpl) bufferProc() {
	logger.Info("Buffer proc started")
	ticker := time.NewTicker(time.Duration(w.service.client.Options().FlushInterval()) * time.Millisecond)
x:
	for {
		select {
		case line := <-w.bufferCh:
			w.writeBuffer = append(w.writeBuffer, line)
			if len(w.writeBuffer) == int(w.service.client.Options().BatchSize()) {
				w.flushBuffer()
			}
		case <-ticker.C:
			w.flushBuffer()
		case <-w.bufferFlush:
			w.flushBuffer()
		case <-w.bufferStop:
			ticker.Stop()
			w.flushBuffer()
			break x
		case buffInfo := <-w.bufferInfoCh:
			buffInfo.writeBuffLen = len(w.bufferInfoCh)
			w.bufferInfoCh <- buffInfo
		}
	}
	logger.Info("Buffer proc finished")
	w.doneCh <- 1
}

func NewPoint(table string, tags map[string]string, values map[string]interface{}, ts time.Time) *Point {
	m := &Point{
		table:     table,
		tags:      nil,
		fields:    nil,
		timestamp: ts,
	}

	if len(tags) > 0 {
		m.tags = make([]*Tag, 0, len(tags))
		for k, v := range tags {
			m.tags = append(m.tags, &Tag{Key: k, Value: v})
		}
	}

	m.fields = make([]*Field, 0, len(fields))
	for k, v := range fields {
		v := convertField(v)
		if v != nil{
			m.fields = append(m.fields, &Field{Key: k, Value: v})
		}	
	}
	// m.SortFields()
	// m.SortTags()
	return m
}

func (p *Point) String() (string, error) {
}

func (m *Point) AddTag(k, v string) *Point {
	for i, tag := range m.tags {
		if k == tag.Key {
			m.tags[i].Value = v
			return m
		}
	}
	m.tags = append(m.tags, &Tag{Key: k, Value: v})
	return m
}

// AddField adds a field to a point.
func (m *Point) AddField(k string, v interface{}) *Point {
	for i, field := range m.fields {
		if k == field.Key {
			m.fields[i].Value = v
			return m
		}
	}
	m.fields = append(m.fields, &Field{Key: k, Value: convertField(v)})
	return m
}

func convertField(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(i)
	case uint64:
		return strconv.FormatUint(v, 10)
	case time.Time:
		return v.String()
	default:
		panic("unsupported type")
	}
}
