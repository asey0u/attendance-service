package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	serverURL string
	http      *http.Client
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: strings.TrimRight(serverURL, "/"),
		http:      &http.Client{},
	}
}

func (c *Client) newRequest(method, path string, body any, token string) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.serverURL+path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

func (c *Client) doRequest(req *http.Request, v interface{}) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	if v != nil {
		return json.Unmarshal(body, v)
	}

	return nil
}

func (c *Client) GetSetupStatus() (bool, error) {
	resp, err := c.http.Get(c.serverURL + "/setup/status")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("server returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var status SetupStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return false, err
	}

	return status.Initialized, nil
}

func (c *Client) CreateAdmin(login, password string) error {
	reqBody, err := json.Marshal(AuthRequest{Login: login, Password: password})
	if err != nil {
		return err
	}

	resp, err := c.http.Post(c.serverURL+"/setup/admin", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return nil
}

func (c *Client) LoginAdmin(login, password string) (string, error) {
	reqBody, err := json.Marshal(AuthRequest{Login: login, Password: password})
	if err != nil {
		return "", err
	}

	resp, err := c.http.Post(c.serverURL+"/login", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var res map[string]string
	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}

	token, ok := res["token"]
	if !ok || token == "" {
		return "", fmt.Errorf("login response did not include token")
	}

	req, err := c.newRequest(http.MethodGet, "/admin/me", nil, token)
	if err != nil {
		return "", err
	}

	var userResp map[string]any
	if err := c.doRequest(req, &userResp); err != nil {
		return "", fmt.Errorf("admin access required: %w", err)
	}

	return token, nil
}

func (c *Client) GetUsers(token string, page int) (*UserListResponse, error) {
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/admin/users?page=%d&size=9", page), nil, token)
	if err != nil {
		return nil, err
	}

	var listResp UserListResponse
	if err := c.doRequest(req, &listResp); err != nil {
		return nil, err
	}

	return &listResp, nil
}

func (c *Client) CreateUser(login, password, role string, employeeID int, token string) error {
	req, err := c.newRequest(http.MethodPost, "/admin/users/create", CreateUserRequest{Login: login, Password: password, Role: role, EmployeeID: employeeID}, token)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}

func (c *Client) ListEmployees(page, size int, token string) (*EmployeeListResponse, error) {
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/admin/employees?page=%d&size=%d", page, size), nil, token)
	if err != nil {
		return nil, err
	}

	var listResp EmployeeListResponse
	if err := c.doRequest(req, &listResp); err != nil {
		return nil, err
	}

	return &listResp, nil
}

func (c *Client) CreateEmployee(firstName, lastName, position, token string) error {
	req, err := c.newRequest(http.MethodPost, "/admin/employees/create", CreateEmployeeRequest{FirstName: firstName, LastName: lastName, Position: position}, token)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}

func (c *Client) GetAttendancesByEmployee(employeeID int, token string) (*AttendanceListResponse, error) {
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/admin/attendances?employee_id=%d", employeeID), nil, token)
	if err != nil {
		return nil, err
	}

	var resp AttendanceListResponse
	if err := c.doRequest(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) CreateAttendance(employeeID int, token string) error {
	req, err := c.newRequest(http.MethodPost, "/admin/attendances/create", map[string]int{"employee_id": employeeID}, token)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}

func (c *Client) DeleteAttendance(attendanceID int, token string) error {
	req, err := c.newRequest(http.MethodPost, fmt.Sprintf("/admin/attendances/delete?attendance_id=%d", attendanceID), nil, token)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}

func (c *Client) GetUserByID(userID int, token string) (*CurrentUserResponse, error) {
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/admin/users/%d", userID), nil, token)
	if err != nil {
		return nil, err
	}

	var resp CurrentUserResponse
	if err := c.doRequest(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) DeleteUser(userID int, token string) error {
	req, err := c.newRequest(http.MethodPost, "/admin/users/delete?id="+strconv.Itoa(userID), nil, token)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}

func (c *Client) UpdateUserRole(userID int, role, token string) error {
	req, err := c.newRequest(http.MethodPost, "/admin/users/update-role?id="+strconv.Itoa(userID), UpdateRoleRequest{Role: role}, token)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}

func (c *Client) ListUsers(page, size int, token string) (*UserListResponse, error) {
	query := fmt.Sprintf("?page=%d&size=%d", page, size)
	req, err := http.NewRequest(http.MethodGet, c.serverURL+"/admin/users"+query, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var listResp UserListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, err
	}

	return &listResp, nil
}
