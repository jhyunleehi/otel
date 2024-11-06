package resty

import (
	"encoding/json"
	"errors"	
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
)

func Test_T1(t *testing.T) {

	collection := "jcollection.json"
	envfile := "jenvironment.json"

	var jdata map[string]interface{}
	var jenv map[string]interface{}
	// Open our jsonFile
	jsonFile, err := os.Open(collection)
	// if we os.Open returns an error then handle it
	if err != nil {
		t.Error(err)
		return
	} 
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		t.Error(err)
		return
	}
	err = json.Unmarshal([]byte(byteValue), &jdata)
	if err != nil {
		t.Error(err)
		return
	}
	//fmt.Printf("[%+v]", jdata)
	envFile, err := os.Open(envfile)
	// if we os.Open returns an error then handle it
	if err != nil {
		t.Error(err)
		return
	}
	defer jsonFile.Close()
	byteValue, err = ioutil.ReadAll(envFile)
	if err != nil {
		t.Error(err)
		return
	}
	err = json.Unmarshal([]byte(byteValue), &jenv)
	if err != nil {
		t.Error(err)
		return
	}

	err = ListRestApi(jdata, jenv)
	if err != nil {
		t.Error(err)
		return
	}
	r := NewResty("")
	err = r.MakeEnv(jenv)
	if err != nil {
		t.Error(err)
	}
	err = r.CallRestApi(jdata)
	if err != nil {
		t.Error(err)
	}
}

func ListRestApi(i interface{}, e interface{}) error {
	t := reflect.TypeOf(i)
	if t.String() != "map[string]interface {}" {
		log.Error(t.String())
		return errors.New("failed convert type")
	}

	item, ok := i.(map[string]interface{})
	if !ok {
		log.Errorf("[%v]", i)
		return errors.New("fail convert type")
	}

	val, ok := item["name"]
	if ok {
		log.Printf("Test:[%s]", val)
	}
	vallist, ok := item["item"]
	if ok {
		for _, v := range vallist.([]interface{}) {
			//log.Debugf("[%d][%v]", k, v)
			err := ListRestApi(v, e)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

/*
func CallRestApi(i interface{}, e interface{}) error {
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
			err := CallRestApi(v, e)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}
	r := NewResty("")
	err := r.RestApi(item)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
*/