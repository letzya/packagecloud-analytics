package main

import (
	"github.com/tomnomnom/linkheader"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type UrlList []EndpointLink
type EndpointLink string
type RepoName string

type downloadDetails struct {
	DownloadedAt time.Time   `json:"downloaded_at"`
	IPAddress    string      `json:"ip_address"`
	UserAgent    string      `json:"user_agent"`
	Source       string      `json:"source"`
	ReadToken    interface{} `json:"read_token"`
}

type Package struct {
	Name               string      `json:"name"`
	DistroVersion      string      `json:"distro_version"`
	CreatedAt          time.Time   `json:"created_at"`
	Version            string      `json:"version"`
	Release            interface{} `json:"release"`
	Epoch              int         `json:"epoch"`
	Scope              interface{} `json:"scope"`
	Private            bool        `json:"private"`
	Type               string      `json:"type"`
	Filename           string      `json:"filename"`
	UploaderName       string      `json:"uploader_name"`
	Indexed            bool        `json:"indexed"`
	Sha256Sum          string      `json:"sha256sum"`
	RepositoryHTMLURL  string      `json:"repository_html_url"`
	PackageURL         string      `json:"package_url"`
	DownloadsDetailURL string      `json:"downloads_detail_url"`
	DownloadsSeriesURL string      `json:"downloads_series_url"`
	DownloadsCountURL  string      `json:"downloads_count_url"`
	PackageHTMLURL     string      `json:"package_html_url"`
	DownloadURL        string      `json:"download_url"`
	PromoteURL         string      `json:"promote_url"`
	DestroyURL         string      `json:"destroy_url"`
}

const GET_METHOD = "GET"
const REPOS_API = "/api/v1/repos"
const TYK_GATEWAY_REPO = "/tyk/tyk-gateway/"
var StartDate string
var EndDate string
var PackageCloudAPIToken string

func main() {

	PackageCloudAPIToken = os.Getenv("PACKAGECLOUD_API_TOKEN")
	if PackageCloudAPIToken == "" {
		fmt.Println("Please set your token as an environment variable. Run in command line the following command with your token in it:")
		fmt.Println("export PACKAGECLOUD_API_TOKEN={your-token}")
		os.Exit(1)
	}
	argsWithoutProg := os.Args[1:]
	StartDate = "start_date=20210301Z" //change to env var
	EndDate = "end_date=20210801Z"

	fmt.Println("Download details for all packages, Page: ", argsWithoutProg)
	StartDate = argsWithoutProg[0]
	EndDate = argsWithoutProg[1]
	//for _, v := range argsWithoutProg[2:] {
	//	fmt.Println("Download details for all packages, Page: ", v)
	//
	//	var page int
	//	var err error
	//	if page, err = strconv.Atoi(v); err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//	if err := GetDownloadsDetailsForPage(TYK_GATEWAY_REPO, page); err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//}

	for _, v := range argsWithoutProg[2:] {
		fmt.Println("Download details for all packages, Page: ", v)

		var page int
		var err error
		if page, err = strconv.Atoi(v); err != nil {
			fmt.Println(err)
			return
		}
		if err := GetDownloadsDetailsForPage(TYK_GATEWAY_REPO, page); err != nil {
			fmt.Println(err)
			return
		}
	}

	//GetDownloadsDetailsForPage(TYK_GATEWAY_REPO, 46)
}

func GetPackagesForPage(repo RepoName, page int) ([]Package, error) {
	const PACKAGE_LIST_FILE = "packages.json"
	packagesAPI := REPOS_API + string(repo) + PACKAGE_LIST_FILE
	if page > 0 {
		packagesAPI += "?page=" + strconv.Itoa(page)
	}

	var linksToPages []linkheader.Link
	body, err := CallPackageCloudApi(packagesAPI, linksToPages)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//fmt.Println("body: ", string(body))
	var packageList []Package
	err = json.Unmarshal(body, &packageList)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return packageList, nil
}

func GetDownloadsDetailsForPage(repo RepoName, page int) error {

	packageList, err := GetPackagesForPage(repo, page)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//
	//var urlList UrlList
	//for _, v := range packageList {
	//	urlList = append(urlList,EndpointLink((v.DownloadsDetailURL)))
	//	//fmt.Printf("%d) CreatedAt: %s\n    download: %s\n\n", i, v.CreatedAt, v.DownloadsDetailURL)
	//}
	//
	//for _, v := range urlList {
	//	//fmt.Printf("%d) url: %s\n",i, v)
	//	if err := GetDownloadDetailsForRepo(v) ; err != nil {
	//		fmt.Println("Error: ", err)
	//		return err
	//	}
	//}

	for _, v := range packageList {
		//fmt.Printf("%d) url: %s\n",i, v)
		//example:
		//https://packagecloud.io/api/v1/repos/tyk/tyk-gateway/package/rpm/el/7/tyk-gateway/x86_64/2.9.4.6/1/stats/downloads/detail.json?start_date=20210418Z&end_date=20210530Z&page=11
		///query params: ?start_date=20210418Z&end_date=20210530Z&page=11
		//api += "?start_date=20210701Z"

		downloadDetailsAPI := v.DownloadsDetailURL + "?" + StartDate + "&" + EndDate
		if err := GetDownloadDetailsForRepo(downloadDetailsAPI); err != nil {
			fmt.Println("Error: ", err)
			return err
		}
	}
	return nil
}

func GetDownloadDetailsForRepo(downloadDetailsApi string ) error {

	downloadDetailsAPI := string(downloadDetailsApi)

	fmt.Printf("inside func GetDownloadDetailsForRepo: downloadDetailsAPI: %s\n", downloadDetailsAPI)

	var linksToPages []linkheader.Link
	body, err := CallPackageCloudApi(downloadDetailsAPI, linksToPages)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//fmt.Printf("url: %s body: %+v\n\n", downloadDetailsAPI, string(body))
	var downloadDetailsList []downloadDetails
	err = json.Unmarshal(body, &downloadDetailsList)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Printf("downloadDetailsAPI: %s, len: %d\n\n", downloadDetailsAPI, len(downloadDetailsList))

	PrintDownloadDetails(downloadDetailsList)
	for _, link := range linksToPages {
		//next last prev first
		const NEXT_PAGE = "next"
		const LAST_PAGE = "last"
//		const FIRST_PAGE = "first"
		if link.Rel == NEXT_PAGE {
			GetDownloadDetailsForRepo(link.URL)
		}
	}

	return nil
}

func PrintDownloadDetails(downloadDetailsList []downloadDetails) {

	for i, v := range downloadDetailsList {
		//t := time.Date(2021, time.July, 1, 00, 00, 00, 0, time.UTC)
		//if v.DownloadedAt.Sub(t) > 0 {
		fmt.Printf("%d) downloadDetails:\n    DownloadedAt: %s\n    IPAddress %s\n    Source: %s\n    UserAgent: %s\n\n",
			i+1, v.DownloadedAt, v.IPAddress, v.Source, v.UserAgent)
	}

}

func CallPackageCloudApi(api string, linksToPages []linkheader.Link) ([]byte, error) {

	const PackageCloudDomain = "https://packagecloud.io"

	api = PackageCloudDomain + api

	//	fmt.Println("api to call: ", api)
	client := &http.Client{}
	req, err := http.NewRequest(GET_METHOD, api, nil)

	if err != nil {
		return nil, fmt.Errorf("Error making request\napi: %s\nerror: %s", api, err)
	}
	req.SetBasicAuth(PackageCloudAPIToken, "")

	//req.Header.Add("Authorization", "Basic "+ApiToken)

	//	fmt.Printf("Calling api: %s\n", api)

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error calling api\napi: %s\nerror: %s", api, err)

	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK, http.StatusCreated:

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("Error when reading response:", err)
		}

		total, perPage, _ := GetPaginationData(res.Header, linksToPages)
		fmt.Printf("total: %d per page: %d\n", total, perPage)

		return body, nil

	case http.StatusUnauthorized, http.StatusNotFound:
		return nil, fmt.Errorf("HTTP status: %s", http.StatusText(res.StatusCode))
	case 422: // Unprocessable Entity
		// fill in
	default:
		return nil, fmt.Errorf("unexpected HTTP status: %d", res.StatusCode)
	}

	return nil, fmt.Errorf("unexpected HTTP status: %d", res.StatusCode)
}

func GetPaginationData(headers http.Header, linksToPages []linkheader.Link) (total int, perPage int, err error) {

	total, _ = strconv.Atoi(headers.Get("Total"))
	perPage, _ = strconv.Atoi(headers.Get("Per-Page"))
	linkHeader := headers.Get("Link")
	//fmt.Printf("total: %d perpage: %d, links: %s\n", total, perPage, linksToPages)

	linksToPages = linkheader.Parse(linkHeader)
	//for _, link := range links {
	//	fmt.Printf("URL: %s; Rel: %s\n", link.URL, link.Rel)
	//}
//links: <https://packagecloud.io/api/v1/repos/tyk/tyk-gateway/packages.json?page=1>; rel="first", <https://packagecloud.io/api/v1/repos/tyk/tyk-gateway/packages.json?page=29>; rel="prev", <https://packagecloud.io/api/v1/repos/tyk/tyk-gateway/packages.json?page=48>; rel="last", <https://packagecloud.io/api/v1/repos/tyk/tyk-gateway/packages.json?page=31>; rel="next"

	return total, perPage, nil
}
