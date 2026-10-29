package ipaddressrangecontainer

import (
	"context"
	"fmt"

	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

type ContainerRecord = cato.ContainerList_Container_List_Containers

// DownloadRecord contains the identity and encoded member file returned by the API.
type DownloadRecord = cato.ContainerIpAddressRangeDownloadFile_Container_IPAddressRange_DownloadFile

// ResourceClient is the narrow account-scoped container API contract.
type ResourceClient interface {
	Find(context.Context, string, string) (*cato.ContainerList_Container_List_Containers, error)
	CreateInline(context.Context, string, m.CreateIPAddressRangeContainerFromListInput) (*cato.ContainerList_Container_List_Containers, error)
	CreateURL(context.Context, string, m.CreateIPAddressRangeContainerFromURLInput) (*cato.ContainerList_Container_List_Containers, error)
	UpdateInline(context.Context, string, m.UpdateIPAddressRangeContainerFromListInput) (*cato.ContainerList_Container_List_Containers, error)
	UpdateURL(context.Context, string, m.UpdateIPAddressRangeContainerFromURLInput) (*cato.ContainerList_Container_List_Containers, error)
	Download(context.Context, string, string) (*cato.ContainerIpAddressRangeDownloadFile_Container_IPAddressRange_DownloadFile, error)
	Delete(context.Context, string, string) (*cato.ContainerList_Container_List_Containers, error)
}
type sdkClient struct{ client *catoclient.Service }

const metadataFields = `id name description size syncData {
fileType url timeInterval timeUnit notifications {
mailingList{
id name
} subscriptionGroup{
id name
} webhook{
id name
}
}
}`

func (c *sdkClient) Find(ctx context.Context, account, id string) (*ContainerRecord, error) {
	// Exact, type-filtered list has reliable empty-list semantics, unlike the nullable search resolver.
	var result cato.ContainerList
	doc := `query containerList($accountId:ID!,
$containerSearchInput:ContainerSearchInput!){
container(accountId:$accountId){
list(input:$containerSearchInput){
containers{
` + metadataFields + `}}}}`
	err := c.client.V2().Client.Post(ctx,
		"containerList",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"containerSearchInput": m.ContainerSearchInput{Refs: []*m.ContainerRefInput{{By: m.ObjectRefByID,
				Input: id}},
				Types: []m.ContainerType{m.ContainerTypeIPRange}}})
	if err != nil {
		return nil,
			err
	}
	list := result.GetContainer().GetList()
	if list == nil ||
		list.Containers == nil {
		return nil,
			fmt.Errorf("API did not return a container list")
	}
	if len(list.Containers) == 0 {
		return nil,
			nil
	}
	if len(list.Containers) != 1 ||
		list.Containers[0] == nil ||
		list.Containers[0].ID != id {
		return nil,
			fmt.Errorf("API did not confirm the requested container identity")
	}
	return list.Containers[0],
		nil
}
func (c *sdkClient) CreateInline(ctx context.Context,
	account string,
	in m.CreateIPAddressRangeContainerFromListInput) (*ContainerRecord,
	error) {
	var result struct {
		Container struct {
			IPAddressRange struct {
				Result struct {
					Container *ContainerRecord `json:"container" graphql:"container"`
				} `json:"createFromList" graphql:"createFromList"`
			} `json:"ipAddressRange" graphql:"ipAddressRange"`
		} `json:"container" graphql:"container"`
	}
	doc := `mutation containerIpAddressRangeCreateFromList($accountId:ID!,
$createIpAddressRangeContainerFromListInput:CreateIpAddressRangeContainerFromListInput!){
container(accountId:$accountId){
ipAddressRange{
createFromList(input:$createIpAddressRangeContainerFromListInput){
container{
` + metadataFields + `}}}}}`
	err := c.client.V2().Client.Post(ctx,
		"containerIpAddressRangeCreateFromList",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"createIpAddressRangeContainerFromListInput": in})
	return result.Container.IPAddressRange.Result.Container,
		err
}
func (c *sdkClient) CreateURL(ctx context.Context,
	account string,
	in m.CreateIPAddressRangeContainerFromURLInput) (*ContainerRecord,
	error) {
	var result struct {
		Container struct {
			IPAddressRange struct {
				Result struct {
					Container *ContainerRecord `json:"container" graphql:"container"`
				} `json:"createFromURL" graphql:"createFromURL"`
			} `json:"ipAddressRange" graphql:"ipAddressRange"`
		} `json:"container" graphql:"container"`
	}
	doc := `mutation containerIpAddressRangeCreateFromURL($accountId:ID!,
$createIpAddressRangeContainerFromUrlInput:CreateIpAddressRangeContainerFromUrlInput!){
container(accountId:$accountId){
ipAddressRange{
createFromURL(input:$createIpAddressRangeContainerFromUrlInput){
container{
` + metadataFields + `}}}}}`
	err := c.client.V2().Client.Post(ctx,
		"containerIpAddressRangeCreateFromURL",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"createIpAddressRangeContainerFromUrlInput": in})
	return result.Container.IPAddressRange.Result.Container,
		err
}
func (c *sdkClient) UpdateInline(ctx context.Context,
	account string,
	in m.UpdateIPAddressRangeContainerFromListInput) (*ContainerRecord,
	error) {
	var result struct {
		Container struct {
			IPAddressRange struct {
				Result struct {
					Container *ContainerRecord `json:"container" graphql:"container"`
				} `json:"updateFromList" graphql:"updateFromList"`
			} `json:"ipAddressRange" graphql:"ipAddressRange"`
		} `json:"container" graphql:"container"`
	}
	doc := `mutation containerIpAddressRangeUpdateFromList($accountId:ID!,
$updateIpAddressRangeContainerFromListInput:UpdateIpAddressRangeContainerFromListInput!){
container(accountId:$accountId){
ipAddressRange{
updateFromList(input:$updateIpAddressRangeContainerFromListInput){
container{
` + metadataFields + `}}}}}`
	err := c.client.V2().Client.Post(ctx,
		"containerIpAddressRangeUpdateFromList",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"updateIpAddressRangeContainerFromListInput": map[string]interface{}{"description": in.Description,
				"ref":    in.Ref,
				"values": in.Values}})
	return result.Container.IPAddressRange.Result.Container,
		err
}
func (c *sdkClient) UpdateURL(ctx context.Context,
	account string,
	in m.UpdateIPAddressRangeContainerFromURLInput) (*ContainerRecord,
	error) {
	var result struct {
		Container struct {
			IPAddressRange struct {
				Result struct {
					Container *ContainerRecord `json:"container" graphql:"container"`
				} `json:"updateFromURL" graphql:"updateFromURL"`
			} `json:"ipAddressRange" graphql:"ipAddressRange"`
		} `json:"container" graphql:"container"`
	}
	doc := `mutation containerIpAddressRangeUpdateFromURL($accountId:ID!,
$updateIpAddressRangeContainerFromUrlInput:UpdateIpAddressRangeContainerFromUrlInput!){
container(accountId:$accountId){
ipAddressRange{
updateFromURL(input:$updateIpAddressRangeContainerFromUrlInput){
container{
` + metadataFields + `}}}}}`
	err := c.client.V2().Client.Post(ctx,
		"containerIpAddressRangeUpdateFromURL",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"updateIpAddressRangeContainerFromUrlInput": in})
	return result.Container.IPAddressRange.Result.Container,
		err
}
func (c *sdkClient) Download(ctx context.Context, account, id string) (*DownloadRecord, error) {
	var result struct {
		Container struct {
			IPAddressRange struct {
				Download *DownloadRecord `json:"downloadFile" graphql:"downloadFile"`
			} `json:"ipAddressRange" graphql:"ipAddressRange"`
		} `json:"container" graphql:"container"`
	}
	doc := `query containerIpAddressRangeDownloadFile($accountId:ID!,
$downloadIpAddressRangeContainerFileInput:DownloadIpAddressRangeContainerFileInput!){
container(accountId:$accountId){
ipAddressRange{
downloadFile(input:$downloadIpAddressRangeContainerFileInput){
id name encodedFile
}
}
}
}`
	err := c.client.V2().Client.Post(ctx,
		"containerIpAddressRangeDownloadFile",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"downloadIpAddressRangeContainerFileInput": m.DownloadIPAddressRangeContainerFileInput{By: m.ObjectRefByID,
				Input: id}})
	return result.Container.IPAddressRange.Download,
		err
}
func (c *sdkClient) Delete(ctx context.Context, account, id string) (*ContainerRecord, error) {
	var result struct {
		Container struct {
			Delete struct {
				Container *ContainerRecord `json:"container" graphql:"container"`
			} `json:"delete" graphql:"delete"`
		} `json:"container" graphql:"container"`
	}
	doc := `mutation containerDelete($accountId:ID!,
$deleteContainerInput:DeleteContainerInput!){
container(accountId:$accountId){
delete(input:$deleteContainerInput){
container{
id name size description
}
}
}
}`
	err := c.client.V2().Client.Post(ctx,
		"containerDelete",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"deleteContainerInput": m.DeleteContainerInput{Ref: &m.ContainerRefInput{By: m.ObjectRefByID,
				Input: id}}})
	return result.Container.Delete.Container,
		err
}
