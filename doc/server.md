### 1. N/A

1. route definition

- Url: /node/list
- Method: GET
- Request: `ListNodeReq`
- Response: `ListNodeResp`

2. request definition



```golang
type ListNodeReq struct {
	PopID string `form:"popid"`
	Type int `form:"type"`
	Start int `form:"start"`
	End int `form:"end"`
}
```


3. response definition



```golang
type ListNodeResp struct {
	Nodes []Node `json:"nodes"`
	Total int `json:"total"`
}
```

### 2. N/A

1. route definition

- Url: /node/pop
- Method: GET
- Request: `GetNodePopReq`
- Response: `GetNodePopResp`

2. request definition



```golang
type GetNodePopReq struct {
	NodeId string `form:"nodeid"`
}
```


3. response definition



```golang
type GetNodePopResp struct {
	ServerURL string `json:"server_url"`
	AccessToken string `json:"access_token"`
}
```

### 3. N/A

1. route definition

- Url: /pops
- Method: GET
- Request: `-`
- Response: `GetPopsResp`

2. request definition



3. response definition



```golang
type GetPopsResp struct {
	Pops []Pop 
}
```

### 4. N/A

1. route definition

- Url: /user/create
- Method: POST
- Request: `CreateUserReq`
- Response: `CreateUserResp`

2. request definition



```golang
type CreateUserReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
	PopId string `json:"pop_id"`
	TrafficLimit *TrafficLimit `json:"traffic_limit,optional"`
	Route *Route `json:"route,optional"`
}
```


3. response definition



```golang
type CreateUserResp struct {
	UserName string `json:"user_name"`
	PopId string `json:"pod_id"`
	TrafficLimit *TrafficLimit `json:"traffic_limit"`
	Route *Route `json:"route"`
	NodeIP string `json:"node_ip"`
}
```

### 5. N/A

1. route definition

- Url: /user/delete
- Method: POST
- Request: `DeleteUserReq`
- Response: `UserOperationResp`

2. request definition



```golang
type DeleteUserReq struct {
	UserName string `json:"user_name"`
}
```


3. response definition



```golang
type UserOperationResp struct {
	Success bool `json:"success"`
	ErrMsg string `json:"err_msg"`
}
```

### 6. N/A

1. route definition

- Url: /user/get
- Method: GET
- Request: `GetUserReq`
- Response: `GetUserResp`

2. request definition



```golang
type GetUserReq struct {
	UserName string `form:"username"`
}
```


3. response definition



```golang
type GetUserResp struct {
	UserName string `json:"user_name"`
	PopId string `json:"pod_id"`
	TrafficLimit *TrafficLimit `json:"traffic_limit"`
	Route *Route `json:"route"`
	NodeIP string `json:"node_ip"`
}
```

### 7. N/A

1. route definition

- Url: /user/list
- Method: GET
- Request: `ListUserReq`
- Response: `ListUserResp`

2. request definition



```golang
type ListUserReq struct {
	PopID string `form:"popid"`
	Start int `form:"start"`
	End int `form:"end"`
}
```


3. response definition



```golang
type ListUserResp struct {
	Users []*User `json:"users"`
	Total int `json:"total"`
}
```

### 8. N/A

1. route definition

- Url: /user/modify
- Method: POST
- Request: `ModifyUserReq`
- Response: `UserOperationResp`

2. request definition



```golang
type ModifyUserReq struct {
	UserName string `json:"user_name"`
	PopId string `json:"pod_id"`
	TrafficLimit *TrafficLimit `json:"total_traffic"`
	Route *Route `json:"route"`
}
```


3. response definition



```golang
type UserOperationResp struct {
	Success bool `json:"success"`
	ErrMsg string `json:"err_msg"`
}
```

### 9. N/A

1. route definition

- Url: /user/password/modify
- Method: POST
- Request: `ModifyUserPasswordReq`
- Response: `UserOperationResp`

2. request definition



```golang
type ModifyUserPasswordReq struct {
	UserName string `json:"user_name"`
	NewPassword string `json:"new_password"`
}
```


3. response definition



```golang
type UserOperationResp struct {
	Success bool `json:"success"`
	ErrMsg string `json:"err_msg"`
}
```

### 10. N/A

1. route definition

- Url: /user/routenode/switch
- Method: POST
- Request: `SwitchUserRouteNodeReq`
- Response: `UserOperationResp`

2. request definition



```golang
type SwitchUserRouteNodeReq struct {
	UserName string `json:"user_name"`
	NodeId string `json:"node_id, optional"`
}
```


3. response definition



```golang
type UserOperationResp struct {
	Success bool `json:"success"`
	ErrMsg string `json:"err_msg"`
}
```

### 11. N/A

1. route definition

- Url: /user/startorstop
- Method: POST
- Request: `StartOrStopUserReq`
- Response: `UserOperationResp`

2. request definition



```golang
type StartOrStopUserReq struct {
	UserName string `json:"user_name"`
	Action string `json:"actoin"`
}
```


3. response definition



```golang
type UserOperationResp struct {
	Success bool `json:"success"`
	ErrMsg string `json:"err_msg"`
}
```

### 12. N/A

1. route definition

- Url: /ws/node
- Method: GET
- Request: `-`
- Response: `-`

2. request definition



3. response definition


