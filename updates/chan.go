package updates

type channel struct {
	c         chan interface{}
	activated bool
}

/*Chan .*/
type Chan interface {
	Get() interface{}
	Push(interface{})
}

/*NewChan .*/
func NewChan(activated bool) Chan {
	return &channel{
		c:         make(chan interface{}),
		activated: activated,
	}
}

func (ch *channel) Get() interface{} {
	if ch.activated {
		msg, ok := <-ch.c
		if ok {
			return msg
		}
	}
	return nil
}

func (ch *channel) Push(e interface{}) {
	if ch.activated {
		go func() { ch.c <- e }()
	}
}
