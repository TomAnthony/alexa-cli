package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the Alexa API client
type Client struct {
	httpClient   *http.Client
	cookies      string
	csrf         string
	amazonDomain string // e.g., "amazon.com"
	customerID   string
}

// NewClient creates a new Alexa API client
func NewClient(refreshToken, amazonDomain string) (*Client, error) {
	client := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		amazonDomain: amazonDomain,
	}

	// Exchange refresh token for cookies
	if err := client.authenticate(refreshToken); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return client, nil
}

// authenticate exchanges a refresh token for session cookies
func (c *Client) authenticate(refreshToken string) error {
	// Amazon token exchange endpoint
	authURL := "https://api.amazon.com/ap/exchangetoken/cookies"

	data := url.Values{}
	data.Set("app_name", "Amazon Alexa")
	data.Set("requested_token_type", "auth_cookies")
	data.Set("source_token_type", "refresh_token")
	data.Set("source_token", refreshToken)
	data.Set("domain", "."+c.amazonDomain)

	req, err := http.NewRequest("POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("x-amzn-identity-auth-domain", "api."+c.amazonDomain)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Response struct {
			Tokens struct {
				Cookies map[string][]struct {
					Name  string `json:"Name"`
					Value string `json:"Value"`
				} `json:"cookies"`
			} `json:"tokens"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	// Build cookie string from response
	var cookieParts []string
	for domain, cookies := range result.Response.Tokens.Cookies {
		_ = domain // cookies are keyed by domain
		for _, cookie := range cookies {
			cookieParts = append(cookieParts, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
		}
	}
	c.cookies = strings.Join(cookieParts, "; ")

	if c.cookies == "" {
		return fmt.Errorf("no cookies received from token exchange")
	}

	// Get CSRF token
	if err := c.fetchCSRF(); err != nil {
		return fmt.Errorf("failed to get CSRF token: %w", err)
	}

	return nil
}

// fetchCSRF retrieves the CSRF token from Amazon
func (c *Client) fetchCSRF() error {
	// Use the language API endpoint which returns CSRF as a cookie
	csrfURL := fmt.Sprintf("https://alexa.%s/api/language", c.amazonDomain)

	req, err := http.NewRequest("GET", csrfURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Cookie", c.cookies)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Extract csrf from response cookies
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrf" {
			c.csrf = cookie.Value
			// Also add csrf to our cookie string for future requests
			c.cookies = c.cookies + "; csrf=" + cookie.Value
			return nil
		}
	}

	// Try to extract from existing cookies (may already be present)
	for _, part := range strings.Split(c.cookies, "; ") {
		if strings.HasPrefix(part, "csrf=") {
			c.csrf = strings.TrimPrefix(part, "csrf=")
			return nil
		}
	}

	return fmt.Errorf("CSRF token not found")
}

// baseURL returns the Alexa API base URL
func (c *Client) baseURL() string {
	// pitangui for US, layla for EU
	if c.amazonDomain == "amazon.com" {
		return "https://pitangui.amazon.com"
	}
	return "https://layla.amazon.com"
}

// alexaURL returns the alexa.amazon.com base URL
func (c *Client) alexaURL() string {
	return fmt.Sprintf("https://alexa.%s", c.amazonDomain)
}

// request makes an authenticated request to the Alexa API (pitangui/layla)
func (c *Client) request(method, endpoint string, body interface{}) ([]byte, error) {
	return c.doRequest(c.baseURL(), method, endpoint, body)
}

// requestAlexa makes an authenticated request to alexa.amazon.com
func (c *Client) requestAlexa(method, endpoint string, body interface{}) ([]byte, error) {
	return c.doRequest(c.alexaURL(), method, endpoint, body)
}

// doRequest makes an authenticated request to the specified base URL
func (c *Client) doRequest(baseURL, method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	fullURL := baseURL + endpoint
	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cookie", c.cookies)
	req.Header.Set("csrf", c.csrf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Device represents an Alexa device
type Device struct {
	AccountName               string `json:"accountName"`
	SerialNumber              string `json:"serialNumber"`
	DeviceType                string `json:"deviceType"`
	DeviceFamily              string `json:"deviceFamily"`
	DeviceOwnerCustomerID     string `json:"deviceOwnerCustomerId"`
	Online                    bool   `json:"online"`
	Capabilities              []string `json:"capabilities"`
}

// GetDevices returns all Alexa devices
func (c *Client) GetDevices() ([]Device, error) {
	data, err := c.request("GET", "/api/devices-v2/device?cached=true", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Devices []Device `json:"devices"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse devices: %w", err)
	}

	// Store customer ID from first device
	if len(result.Devices) > 0 {
		c.customerID = result.Devices[0].DeviceOwnerCustomerID
	}

	return result.Devices, nil
}

// SequenceCommand sends a sequence command to a device
func (c *Client) SequenceCommand(device *Device, command string) error {
	// Ensure we have customer ID
	if c.customerID == "" {
		c.customerID = device.DeviceOwnerCustomerID
	}

	// Parse command type and build the appropriate payload
	var sequenceJson string

	switch {
	case strings.HasPrefix(command, "speak:"):
		text := strings.TrimPrefix(command, "speak:")
		text = strings.Trim(text, "'\"")

		sequenceJson = fmt.Sprintf(`{
			"@type": "com.amazon.alexa.behaviors.model.Sequence",
			"startNode": {
				"@type": "com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode",
				"type": "Alexa.Speak",
				"operationPayload": {
					"deviceType": "%s",
					"deviceSerialNumber": "%s",
					"customerId": "%s",
					"locale": "en-US",
					"textToSpeak": %s
				}
			}
		}`, device.DeviceType, device.SerialNumber, c.customerID, mustJSON(text))

	case strings.HasPrefix(command, "announcement:"):
		text := strings.TrimPrefix(command, "announcement:")
		text = strings.Trim(text, "'\"")

		sequenceJson = fmt.Sprintf(`{
			"@type": "com.amazon.alexa.behaviors.model.Sequence",
			"startNode": {
				"@type": "com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode",
				"type": "AlexaAnnouncement",
				"operationPayload": {
					"expireAfter": "PT5S",
					"content": [{
						"locale": "en-US",
						"display": {"title": "Announcement", "body": %s},
						"speak": {"type": "text", "value": %s}
					}],
					"target": {
						"customerId": "%s"
					}
				}
			}
		}`, mustJSON(text), mustJSON(text), c.customerID)

	case strings.HasPrefix(command, "textcommand:"):
		text := strings.TrimPrefix(command, "textcommand:")
		text = strings.Trim(text, "'\"")

		sequenceJson = fmt.Sprintf(`{
			"@type": "com.amazon.alexa.behaviors.model.Sequence",
			"startNode": {
				"@type": "com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode",
				"type": "Alexa.TextCommand",
				"skillId": "amzn1.ask.1p.tellalexa",
				"operationPayload": {
					"deviceType": "%s",
					"deviceSerialNumber": "%s",
					"customerId": "%s",
					"text": %s
				}
			}
		}`, device.DeviceType, device.SerialNumber, c.customerID, mustJSON(text))

	case strings.HasPrefix(command, "automation:"):
		routineName := strings.TrimPrefix(command, "automation:")
		routineName = strings.Trim(routineName, "'\"")
		return c.ExecuteRoutine(device, routineName)

	default:
		return fmt.Errorf("unknown command type: %s", command)
	}

	payload := map[string]interface{}{
		"behaviorId":   "PREVIEW",
		"sequenceJson": sequenceJson,
		"status":       "ENABLED",
	}

	_, err := c.request("POST", "/api/behaviors/preview", payload)
	return err
}

// ExecuteRoutine runs an Alexa routine by name
func (c *Client) ExecuteRoutine(device *Device, routineName string) error {
	// First, get the list of routines
	routines, err := c.GetRoutines()
	if err != nil {
		return fmt.Errorf("failed to get routines: %w", err)
	}

	// Find matching routine
	var targetRoutine *Routine
	nameLower := strings.ToLower(routineName)
	for i, r := range routines {
		if strings.ToLower(r.Name) == nameLower {
			targetRoutine = &routines[i]
			break
		}
	}

	if targetRoutine == nil {
		return fmt.Errorf("routine '%s' not found", routineName)
	}

	// Execute the routine
	payload := map[string]interface{}{
		"behaviorId": targetRoutine.AutomationID,
		"sequenceJson": targetRoutine.Sequence,
		"status": "ENABLED",
	}

	_, err = c.request("POST", "/api/behaviors/preview", payload)
	return err
}

// Routine represents an Alexa routine
type Routine struct {
	AutomationID string `json:"automationId"`
	Name         string `json:"name"`
	Sequence     string `json:"sequence"`
}

// GetRoutines returns all Alexa routines
func (c *Client) GetRoutines() ([]Routine, error) {
	// Routines are on alexa.amazon.com, not pitangui
	data, err := c.requestAlexa("GET", "/api/behaviors/automations", nil)
	if err != nil {
		return nil, err
	}

	var rawRoutines []struct {
		AutomationID string `json:"automationId"`
		Name         string `json:"name"`
		Sequence     json.RawMessage `json:"sequence"`
	}

	if err := json.Unmarshal(data, &rawRoutines); err != nil {
		return nil, fmt.Errorf("failed to parse routines: %w", err)
	}

	routines := make([]Routine, len(rawRoutines))
	for i, r := range rawRoutines {
		routines[i] = Routine{
			AutomationID: r.AutomationID,
			Name:         r.Name,
			Sequence:     string(r.Sequence),
		}
	}

	return routines, nil
}

// SmartHomeDevice represents a smart home device
type SmartHomeDevice struct {
	EntityID     string `json:"entityId"`
	ApplianceID  string `json:"applianceId"`
	Name         string `json:"friendlyName"`
	Description  string `json:"friendlyDescription"`
	Type         string `json:"applianceTypes"`
	Reachable    bool   `json:"isReachable"`
}

// GetSmartHomeDevices returns all smart home devices
func (c *Client) GetSmartHomeDevices() ([]SmartHomeDevice, error) {
	data, err := c.request("GET", "/api/phoenix", nil)
	if err != nil {
		return nil, err
	}

	// Phoenix response has nested structure
	var result struct {
		NetworkDetail []struct {
			ApplianceDetails map[string]SmartHomeDevice `json:"applianceDetails"`
		} `json:"networkDetail"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse smart home devices: %w", err)
	}

	var devices []SmartHomeDevice
	for _, network := range result.NetworkDetail {
		for _, device := range network.ApplianceDetails {
			devices = append(devices, device)
		}
	}

	return devices, nil
}

// ControlSmartHome controls a smart home device
func (c *Client) ControlSmartHome(entityID string, action string, value interface{}) error {
	var payload map[string]interface{}

	switch action {
	case "on", "turnOn":
		payload = map[string]interface{}{
			"controlRequests": []map[string]interface{}{
				{
					"entityId":  entityID,
					"entityType": "APPLIANCE",
					"parameters": map[string]interface{}{
						"action": "turnOn",
					},
				},
			},
		}
	case "off", "turnOff":
		payload = map[string]interface{}{
			"controlRequests": []map[string]interface{}{
				{
					"entityId":  entityID,
					"entityType": "APPLIANCE",
					"parameters": map[string]interface{}{
						"action": "turnOff",
					},
				},
			},
		}
	case "brightness":
		payload = map[string]interface{}{
			"controlRequests": []map[string]interface{}{
				{
					"entityId":  entityID,
					"entityType": "APPLIANCE",
					"parameters": map[string]interface{}{
						"action": "setBrightness",
						"brightness": value,
					},
				},
			},
		}
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	_, err := c.request("PUT", "/api/phoenix/state", payload)
	return err
}

func mustJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
