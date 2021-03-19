package rigger

import "github.com/AsynkronIT/protoactor-go/actor"

// 根据error生成一个rigger.Error
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{ErrStr: err.Error()}
}

// 根据字符串生成一个rigger.Error
func FromString(str string) *Error {
	return &Error{ErrStr: str}
}

// 将一个类型为error的接口转换成error
func ToError(maybeError interface{}) error {
	if maybeError == nil {
		return nil
	} else {
		return maybeError.(error)
	}
}

/*
如果 future的结果是error, 则此函数可以快捷的取出future的错误,如果有错,则返回error,否则返回nil
*/
func FetchFutureError(future *actor.Future) error {
	if ret, err := future.Result(); err != nil {
		return err
	} else {
		return ToError(ret)
	}
}

