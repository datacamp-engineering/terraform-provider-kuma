package kumaclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// const (
// 	BaseURLV1 = "http://localhost:5681"
// )

type KumaClient struct {
	BaseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(host string, apiKey string) (c *KumaClient, err error) {
	client := &KumaClient{
		BaseURL: host,
		apiKey:  apiKey,
		http: &http.Client{
			Timeout: time.Minute,
		},
	}

	return client, nil
}

func (c KumaClient) GetMesh(name string) (m Mesh, e error) {
	fmt.Println("1. Performing Http Get...")
	resp, err := c.http.Get(fmt.Sprintf("%s/meshes/%s", c.BaseURL, name))
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to string
	bodyString := string(bodyBytes)
	fmt.Println("API Response as String:\n" + bodyString)

	// Convert response body to Todo struct
	var mesh Mesh
	json.Unmarshal(bodyBytes, &mesh)
	fmt.Printf("API Response as struct %+v\n", mesh)
	return mesh, nil
}

func (c KumaClient) UpsertMesh(mesh Mesh) error {
	json, err := json.Marshal(mesh)
	if err != nil {
		panic(err)
	}

	url := fmt.Sprintf("%s/meshes/%s", c.BaseURL, mesh.Name)
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(json))
	if err != nil {
		log.Fatalln(err)
	}

	request.Header["Content-Type"] = []string{"application/json"}

	resp, err := c.http.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to string
	bodyString := string(bodyBytes)
	fmt.Println(resp.Status)
	fmt.Println("API Response as String:\n" + bodyString)
	return nil
}

func (c KumaClient) DeleteMesh(mesh Mesh) error {
	return c.DeleteMeshByName(mesh.Name)
}

func (c KumaClient) DeleteMeshByName(name string) error {
	url := fmt.Sprintf("%s/meshes/%s", c.BaseURL, name)
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	request.Header["Content-Type"] = []string{"application/json"}

	resp, err := c.http.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to string
	bodyString := string(bodyBytes)
	fmt.Println(resp.Status)
	fmt.Println("API Response as String:\n" + bodyString)
	return nil
}

func (c KumaClient) GetDataPlane(mesh string, dpName string) (d DataPlane, e error) {
	url := fmt.Sprintf("%s/meshes/%s/dataplanes/%s", c.BaseURL, mesh, dpName)
	fmt.Println("1. Performing Http Get...")
	resp, err := c.http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to string
	bodyString := string(bodyBytes)
	fmt.Println("API Response as String:\n" + bodyString)

	// Convert response body to Todo struct
	var dataplane DataPlane
	json.Unmarshal(bodyBytes, &dataplane)
	fmt.Printf("API Response as struct %+v\n", dataplane)
	return dataplane, nil
}

func (c KumaClient) UpsertDataplane(dataplane DataPlane) error {
	json, err := json.Marshal(dataplane)
	if err != nil {
		panic(err)
	}

	url := fmt.Sprintf("%s/meshes/%s/dataplanes/%s", c.BaseURL, dataplane.Mesh, dataplane.Name)
	fmt.Println(url)
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(json))
	if err != nil {
		log.Fatalln(err)
	}

	request.Header["Content-Type"] = []string{"application/json"}

	resp, err := c.http.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to string
	bodyString := string(bodyBytes)
	fmt.Println(resp.Status)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("%s: %s", resp.Status, bodyString)
	}

	fmt.Println("API Response as String:\n" + bodyString)
	return nil
}

func (c KumaClient) DeleteDataplane(dataplane DataPlane) error {
	return c.DeleteDataplaneByName(dataplane.Mesh, dataplane.Name)
}

func (c KumaClient) DeleteDataplaneByName(name string, dpName string) error {
	url := fmt.Sprintf("%s/meshes/%s/dataplanes/%s", c.BaseURL, name, dpName)
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	request.Header["Content-Type"] = []string{"application/json"}

	resp, err := c.http.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to string
	bodyString := string(bodyBytes)
	fmt.Println(resp.Status)
	fmt.Println("API Response as String:\n" + bodyString)
	return nil
}
