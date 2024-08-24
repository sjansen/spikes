package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Credentials struct {
	cfg *oauth2.Config
	tok *oauth2.Token

	token_filename string
}

func NewCredentials() (*Credentials, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("unable to get user config dir: %v", err)
	}

	c := &Credentials{
		token_filename: filepath.Join(dir, "inbox-status", "token.json"),
	}

	if err := c.setOauth2Config(); err != nil {
		return nil, err
	}

	if err := c.setToken(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Credentials) GetClient(ctx context.Context) *http.Client {
	return c.cfg.Client(ctx, c.tok)
}

func (c *Credentials) setOauth2Config() error {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token file.
	c.cfg, err = google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	return nil
}

func (c *Credentials) saveToken() error {
	dir := filepath.Dir(c.token_filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("unable to create config directory: %v", err)
		}
	}

	f, err := os.OpenFile(c.token_filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(c.tok)
}

func (c *Credentials) setToken() error {
	err := c.setTokenFromFile()
	if err != nil {
		if err := c.setTokenFromWeb(); err != nil {
			return err
		}
		if err := c.saveToken(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Credentials) setTokenFromFile() error {
	f, err := os.Open(c.token_filename)
	if err != nil {
		return err
	}
	defer f.Close()

	c.tok = &oauth2.Token{}
	err = json.NewDecoder(f).Decode(c.tok)
	return err
}

func (c *Credentials) setTokenFromWeb() error {
	authURL := c.cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then copy & paste the authorization code: \n%v\n\ncode=", authURL)

	var authCode string
	_, err := fmt.Scan(&authCode)
	if err != nil {
		return fmt.Errorf("unable to read authorization code: %v", err)
	}

	c.tok, err = c.cfg.Exchange(context.TODO(), authCode)
	if err != nil {
		return fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	return nil
}

func die(e error) {
	fmt.Fprintln(os.Stderr, e)
	os.Exit(1)
}

func main() {
	ctx := context.Background()
	c, err := NewCredentials()
	if err != nil {
		die(err)
	}

	client := c.GetClient(ctx)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		die(fmt.Errorf("unable to retrieve Gmail client: %v", err))
	}

	resp, err := srv.Users.Labels.List("me").Do()
	if err != nil {
		die(fmt.Errorf("unable to retrieve labels: %v", err))
	}

	if len(resp.Labels) == 0 {
		fmt.Println("No labels found.")
		return
	}

	fmt.Println("Labels:")
	for _, l := range resp.Labels {
		fmt.Printf("- %s\n", l.Name)
	}
}
