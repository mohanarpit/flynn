// Package v2controller provides a client for v2 of the controller API (GraphQL).
package v2controller

import (
	"io"
	"time"

	"github.com/flynn/flynn/controller/client/v1"
	"github.com/flynn/flynn/pkg/httpclient"
	"github.com/flynn/flynn/pkg/stream"
	"github.com/flynn/flynn/router/types"
)

// Client is a client for the v2 of the controller API (GraphQL).
type Client struct {
	*httpclient.Client

	v1client *v1controller.Client
}

func (c *Client) GetCACert() ([]byte, error) {
	c.v1client.GetCACert()
}

func (c *Client) StreamFormations(since *time.Time, output chan<- *ct.ExpandedFormation) (stream.Stream, error) {
	c.v1client.StreamFormations(since, output)
}

func (c *Client) PutDomain(dm *ct.DomainMigration) error {
	c.v1client.PutDomain(dm)
}

func (c *Client) CreateArtifact(artifact *ct.Artifact) error {
	c.v1client.CreateArtifact(artifact)
}

func (c *Client) CreateRelease(release *ct.Release) error {
	c.v1client.CreateRelease(release)
}

func (c *Client) CreateApp(app *ct.App) error {
	c.v1client.CreateApp(app)
}

func (c *Client) UpdateApp(app *ct.App) error {
	c.v1client.UpdateApp(app)
}

func (c *Client) UpdateAppMeta(app *ct.App) error {
	c.v1client.UpdateAppMeta(app)
}

func (c *Client) DeleteApp(appID string) (*ct.AppDeletion, error) {
	c.v1client.DeleteApp(appID)
}

func (c *Client) CreateProvider(provider *ct.Provider) error {
	c.v1client.CreateProvider(provider)
}

func (c *Client) GetProvider(providerID string) (*ct.Provider, error) {
	c.v1client.GetProvider(providerID)
}

func (c *Client) ProvisionResource(req *ct.ResourceReq) (*ct.Resource, error) {
	c.v1client.ProvisionResource(req)
}

func (c *Client) GetResource(providerID, resourceID string) (*ct.Resource, error) {
	c.v1client.GetResource(providerID, resourceID)
}

func (c *Client) ResourceListAll() ([]*ct.Resource, error) {
	c.v1client.ResourceListAll()
}

func (c *Client) ResourceList(providerID string) ([]*ct.Resource, error) {
	c.v1client.ResourceList(providerID)
}

func (c *Client) AddResourceApp(providerID, resourceID, appID string) (*ct.Resource, error) {
	c.v1client.AddResourceApp(providerID, resourceID, appID)
}

func (c *Client) DeleteResourceApp(providerID, resourceID, appID string) (*ct.Resource, error) {
	c.v1client.DeleteResourceApp(providerID, resourceID, appID)
}

func (c *Client) AppResourceList(appID string) ([]*ct.Resource, error) {
	c.v1client.AppResourceList(appID)
}

func (c *Client) PutResource(resource *ct.Resource) error {
	c.v1client.PutResource(resource)
}

func (c *Client) DeleteResource(providerID, resourceID string) (*ct.Resource, error) {
	c.v1client.DeleteResource(providerID, resourceID)
}

func (c *Client) PutFormation(formation *ct.Formation) error {
	c.v1client.PutFormation(formation)
}

func (c *Client) PutJob(job *ct.Job) error {
	c.v1client.PutJob(job)
}

func (c *Client) DeleteJob(appID, jobID string) error {
	c.v1client.DeleteJob(appID, jobID)
}

func (c *Client) SetAppRelease(appID, releaseID string) error {
	c.v1client.SetAppRelease(appID, releaseID)
}

func (c *Client) GetAppRelease(appID string) (*ct.Release, error) {
	c.v1client.GetAppRelease(appID)
}

func (c *Client) RouteList(appID string) ([]*router.Route, error) {
	c.v1client.RouteList(appID)
}

func (c *Client) GetRoute(appID string, routeID string) (*router.Route, error) {
	c.v1client.GetRoute(appID, routeID)
}

func (c *Client) CreateRoute(appID string, route *router.Route) error {
	c.v1client.CreateRoute(appID, route)
}

func (c *Client) UpdateRoute(appID string, routeID string, route *router.Route) error {
	c.v1client.UpdateRoute(appID, routeID, route)
}

func (c *Client) DeleteRoute(appID string, routeID string) error {
	c.v1client.DeleteRoute(appID, routeID)
}

func (c *Client) GetFormation(appID, releaseID string) (*ct.Formation, error) {
	c.v1client.GetFormation(appID, releaseID)
}

func (c *Client) GetExpandedFormation(appID, releaseID string) (*ct.ExpandedFormation, error) {
	c.v1client.GetExpandedFormation(appID, releaseID)
}

func (c *Client) FormationList(appID string) ([]*ct.Formation, error) {
	c.v1client.FormationList(appID)
}

func (c *Client) FormationListActive() ([]*ct.ExpandedFormation, error) {
	c.v1client.FormationListActive()
}

func (c *Client) DeleteFormation(appID, releaseID string) error {
	c.v1client.DeleteFormation(appID, releaseID)
}

func (c *Client) GetRelease(releaseID string) (*ct.Release, error) {
	c.v1client.GetRelease(releaseID)
}

func (c *Client) GetArtifact(artifactID string) (*ct.Artifact, error) {
	c.v1client.GetArtifact(artifactID)
}

func (c *Client) GetApp(appID string) (*ct.App, error) {
	c.v1client.GetApp(appID)
}

func (c *Client) GetAppLog(appID string, options *ct.LogOpts) (io.ReadCloser, error) {
	c.v1client.GetAppLog(appID, options)
}

func (c *Client) StreamAppLog(appID string, options *ct.LogOpts, output chan<- *ct.SSELogChunk) (stream.Stream, error) {
	c.v1client.StreamAppLog(appID, optionts, output)
}

func (c *Client) GetDeployment(deploymentID string) (*ct.Deployment, error) {
	c.v1client.GetDeployment(deploymentID)
}

func (c *Client) CreateDeployment(appID, releaseID string) (*ct.Deployment, error) {
	c.v1client.CreateDeployment(appID, releaseID)
}

func (c *Client) DeploymentList(appID string) ([]*ct.Deployment, error) {
	c.v1client.DeploymentList(appID)
}

func (c *Client) StreamDeployment(d *ct.Deployment, output chan *ct.DeploymentEvent) (stream.Stream, error) {
	c.v1client.StreamDeployment(d, output)
}

func (c *Client) DeployAppRelease(appID, releaseID string, stopWait <-chan struct{}) error {
	c.v1client.DeployAppRelease(appID, releaseID, stopWait)
}

func (c *Client) StreamJobEvents(appID string, output chan *ct.Job) (stream.Stream, error) {
	c.v1client.StreamJobEvents(appID, output)
}

func (c *Client) WatchJobEvents(appID, releaseID string) (ct.JobWatcher, error) {
	c.v1client.WatchJobEvents(appID, releaseID)
}

func (c *Client) StreamEvents(opts ct.StreamEventsOptions, output chan *ct.Event) (stream.Stream, error) {
	c.v1client.StreamEvents(opts)
}

func (c *Client) ListEvents(opts ct.ListEventsOptions) ([]*ct.Event, error) {
	c.v1client.ListEvents(opts)
}

func (c *Client) GetEvent(id int64) (*ct.Event, error) {
	c.v1client.GetEvent(id)
}

func (c *Client) ExpectedScalingEvents(actual, expected map[string]int, releaseProcesses map[string]ct.ProcessType, clusterSize int) ct.JobEvents {
	c.v1client.ExpectedScalingEvents(actual, expected, releaseProcesses, clusterSize)
}

func (c *Client) RunJobAttached(appID string, job *ct.NewJob) (httpclient.ReadWriteCloser, error) {
	c.v1client.RunJobAttached(appID, job)
}

func (c *Client) RunJobDetached(appID string, req *ct.NewJob) (*ct.Job, error) {
	c.v1client.RunJobDetached(appID, req)
}

func (c *Client) GetJob(appID, jobID string) (*ct.Job, error) {
	c.v1client.GetJob(appID, jobID)
}

func (c *Client) JobList(appID string) ([]*ct.Job, error) {
	c.v1client.JobList(appID)
}

func (c *Client) JobListActive() ([]*ct.Job, error) {
	c.v1client.JobListActive()
}

func (c *Client) AppList() ([]*ct.App, error) {
	c.v1client.AppList()
}

func (c *Client) KeyList() ([]*ct.Key, error) {
	c.v1client.KeyList()
}

func (c *Client) ArtifactList() ([]*ct.Artifact, error) {
	c.v1client.ArtifactList()
}

func (c *Client) ReleaseList() ([]*ct.Release, error) {
	c.v1client.ReleaseList()
}

func (c *Client) AppReleaseList(appID string) ([]*ct.Release, error) {
	c.v1client.AppReleaseList(appID)
}

func (c *Client) CreateKey(pubKey string) (*ct.Key, error) {
	c.v1client.CreateKey(pubKey)
}

func (c *Client) GetKey(keyID string) (*ct.Key, error) {
	c.v1client.GetKey(keyID)
}

func (c *Client) DeleteKey(id string) error {
	c.v1client.DeleteKey(id)
}

func (c *Client) ProviderList() ([]*ct.Provider, error) {
	c.v1client.ProviderList()
}

func (c *Client) Backup() (io.ReadCloser, error) {
	c.v1client.Backup()
}

func (c *Client) GetBackupMeta() (*ct.ClusterBackup, error) {
	c.v1client.GetBackupMeta()
}

func (c *Client) DeleteRelease(appID, releaseID string) (*ct.ReleaseDeletion, error) {
	c.v1client.DeleteRelease(appID, releaseID)
}

func (c *Client) ScheduleAppGarbageCollection(appID string) error {
	c.v1client.ScheduleAppGarbageCollection(appID)
}
