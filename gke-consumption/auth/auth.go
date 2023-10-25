// Package auth contains authentication related util functions for e2e test.
package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"

	container "cloud.google.com/go/container/apiv1"
	proto "cloud.google.com/go/container/apiv1/containerpb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// DefaultTokenSource creates an OAuth2 token source of default auth scope
func DefaultTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	ts, err := google.DefaultTokenSource(ctx, container.DefaultAuthScopes()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create default token source: %v", err)
	}
	return ts, nil
}

func clusterCACertPool(c *proto.Cluster) (*x509.CertPool, error) {
	certPEM, err := base64.StdEncoding.DecodeString(c.GetMasterAuth().GetClusterCaCertificate())
	if err != nil {
		return nil, fmt.Errorf("unable to decode CA cert for cluster %s: %v", c.Name, err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certPEM) {
		return nil, fmt.Errorf("failed to append cert from PEM for cluster %s", c.Name)
	}

	return certPool, nil
}

// OAuthTransport creates a transport with OAuth2 token
func OAuthTransport(ctx context.Context, cluster *proto.Cluster) (*oauth2.Transport, error) {
	cp, err := clusterCACertPool(cluster)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: cp,
		},
	}
	ts, err := DefaultTokenSource(ctx)
	if err != nil {
		return nil, err
	}
	otr := &oauth2.Transport{
		Base:   tr,
		Source: ts,
	}
	return otr, nil
}
