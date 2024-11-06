package resty

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	resty "github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
	log.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TimestampFormat: time.RFC3339,
		NoColors:        true,
	})
}

type Resty struct {
	name    string
	client  *resty.Client
	request *resty.Request
	env     map[string]string
	mutex   sync.Mutex
}

func NewResty(name string) *Resty {
	r := Resty{
		name: name,
	}
	r.env = make(map[string]string)
	r.client = resty.New()
	//r.client.SetDebug(true)
	r.client.SetDebug(false)
	r.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	r.client.SetTimeout(1 * time.Minute)
	r.request = r.client.R()
	return &r
}

func (r *Resty) MakeEnv(env interface{}) error {
	item, ok := env.(map[string]interface{})
	if !ok {
		log.Errorf("[%v]", env)
		return errors.New("fail convert type")
	}
	t := reflect.TypeOf(env)
	if t.String() != "map[string]interface {}" {
		log.Error(t.String())
		return errors.New("failed convert type")
	}

	values, ok := item["values"]
	if ok {
		for _, val := range values.([]interface{}) {
			v := val.(map[string]interface{})
			r.env[v["key"].(string)] = v["value"].(string)
		}
	}
	return nil
}

func (r *Resty) CallRestApi(i interface{}) error {
	item, ok := i.(map[string]interface{})
	if !ok {
		log.Errorf("[%v]", i)
		return errors.New("fail convert type")
	}
	t := reflect.TypeOf(i)
	if t.String() != "map[string]interface {}" {
		log.Error(t.String())
		return errors.New("failed convert type")
	}

	val, ok := item["name"]
	if ok {
		log.Printf("TEST: [%s]", val)
	}
	vallist, ok := item["item"]
	if ok {
		for _, v := range vallist.([]interface{}) {
			//log.Debugf("[%d][%v]", k, v)
			err := r.CallRestApi(v)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}
	err := r.RestApi(item)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (r *Resty) DoRestApi(method string, restitem, in, out interface{}) error {
	item := restitem.(map[string]interface{})
	headers, ok := item["header"]
	if ok {
		for _, head := range headers.([]interface{}) {
			h := head.(map[string]interface{})
			r.request = r.request.SetHeader(h["key"].(string), h["value"].(string))
		}
	}
	url, ok := item["url"]
	if !ok {
		msg := fmt.Sprintf("Cannot find url field [%s]", url)
		log.Error(msg)
		return errors.New(msg)
	}
	raw := url.(string)
	log.Debugf("%s", url)

	if in != nil {
		bodystring, err := json.Marshal(in)
		if err != nil {
			log.Error(err)
			return err
		}
		r.request.SetBody(bodystring)
	}

	switch method {
	case "GET":
		res, err := r.request.Get(raw)
		if err != nil {
			log.Error(err)
			return err
		}
		if len(res.Body()) > 0 {
			log.Debugf("%s", string(res.Body()))
			if out != nil {
				err := json.Unmarshal(res.Body(), out)
				if err != nil {
					log.Error(err)
				}
			}
		}
	case "POST":
		res, err := r.request.Post(raw)
		if err != nil {
			log.Error(err)
			return err
		}
		if len(res.Body()) > 0 {
			log.Debugf("%s", string(res.Body()))
			if out != nil {
				err := json.Unmarshal(res.Body(), out)
				if err != nil {
					log.Error(err)
				}
			}
		}
	case "PUT":
		res, err := r.request.Put(raw)
		if err != nil {
			log.Error(err)
			return err
		}
		if len(res.Body()) > 0 {
			log.Debugf("%s", string(res.Body()))
			if out != nil {
				err := json.Unmarshal(res.Body(), out)
				if err != nil {
					log.Error(err)
				}
			}
		}
	case "PATCH":
		res, err := r.request.Patch(raw)
		if err != nil {
			log.Error(err)
			return err
		}
		if len(res.Body()) > 0 {
			log.Debugf("%s", string(res.Body()))
			if out != nil {
				err := json.Unmarshal(res.Body(), out)
				if err != nil {
					log.Error(err)
				}
			}
		}
	case "DELETE":
		res, err := r.request.Delete(raw)
		if err != nil {
			log.Error(err)
			return err
		}
		if len(res.Body()) > 0 {
			log.Debugf("%s", string(res.Body()))
			if out != nil {
				err := json.Unmarshal(res.Body(), out)
				if err != nil {
					log.Error(err)
				}
			}
		}
	default:
	}
	return nil
}

func (r *Resty) RestApi(restitem interface{}) error {
	//log.Debugf("[%+v]", restitem)
	item := restitem.(map[string]interface{})

	name, ok := item["name"]
	if !ok {
		msg := fmt.Sprintf("Cannot find name filed [%s]", name)
		log.Debug(msg)
		return nil
	}
	log.Printf("Test:[%s]", name)

	req, ok := item["request"]
	if !ok {
		//return nil
		msg := fmt.Sprintf("Cannot find request filed [%s]", req)
		log.Debug(msg)
		return nil
	}
	request := req.(map[string]interface{})
	headers, ok := request["header"]
	if ok {
		for _, head := range headers.([]interface{}) {
			h := head.(map[string]interface{})
			r.request = r.request.SetHeader(h["key"].(string), h["value"].(string))
		}
	}
	url, ok := request["url"]
	if !ok {
		msg := fmt.Sprintf("Cannot find url field [%s]", url)
		log.Error(msg)
		return errors.New(msg)
	}
	urlmap := url.(map[string]interface{})
	raw := urlmap["raw"].(string)
	log.Debugf("%s", raw)

	body, ok := request["body"]
	if ok {
		bodymap := body.(map[string]interface{})
		bodystring := bodymap["raw"].(string)
		r.request.SetBody(bodystring)
	}

	method, ok := request["method"].(string)
	if !ok {
		msg := fmt.Sprintf("Cannot find method filed [%s]", method)
		log.Error(msg)
		return errors.New(msg)
	}
	switch method {
	case "GET":
		res, err := r.request.Get(raw)
		if err != nil {
			log.Error(err)
			return err
		}
		if len(res.Body()) > 0 {
			log.Debugf("%s", string(res.Body()))
		}
	case "POST":
		res, err := r.request.Post(raw)
		if err != nil {
			log.Error(err)
			return err
		}
		if len(res.Body()) > 0 {
			log.Debugf("%s", string(res.Body()))
		}
	case "PUT":
		res, err := r.request.Put(raw)
		if err != nil {
			log.Error(err)
			return err
		}
		if len(res.Body()) > 0 {
			log.Debugf("%s", string(res.Body()))
		}
	case "PATCH":
		res, err := r.request.Patch(raw)
		if err != nil {
			log.Error(err)
			return err
		}
		if len(res.Body()) > 0 {
			log.Debugf("%s", string(res.Body()))
		}
	case "DELETE":
		res, err := r.request.Delete(raw)
		if err != nil {
			log.Error(err)
			return err
		}
		if len(res.Body()) > 0 {
			log.Debugf("%s", string(res.Body()))
		}
	default:

	}
	return nil
}
