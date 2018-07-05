# Real Go HTTP Mock

A go project to up http mocks resources.

## How to use

It is necessary to create a json configuration file. Bellow is a example:

```json
{
    "port": 9001,
    "resources": [{
        "path": "/cars",
        "methods": [{
            "name": "A resource to get cars",
            "type": "GET",
            "conversations": [{
                "response": {
                    "content-type": "application/json",
                    "body": "{ \"car\": \"fusca\" }",
                    "code": 200
                }
            }]
        }]
    }]
}
```

We can configure several resource paths to serve. A resource struct has a list of different methods and a path which will serve

```json
{
    "port": 9001,
    "resources": [{
        "path": "/cars",
        "methods": [{
                "name": "A resource to get cars",
                "type": "GET",
                "conversations": [{
                    "response": {
                        "content-type": "application/json",
                        "body": "{ \"car\": \"fusca\" }",
                        "code": 200
                    }
                }]
            },
            {
                "name": "A resource to create cars",
                "type": "POST",
                "conversations": [{
                    "response": {
                        "content-type": "text/plain",
                        "body": "OK",
                        "code": 201
                    }
                }]
            }
        ]
    }]
}
```
We can configure a path parameter and the response depends of the parameter value
```json
{
    "port": 9001,
    "resources": [{
        "path": "/cars/{param}",
        "methods": [{
                "name": "A resource to get cars",
                "type": "GET",
                "conversations": [{
                    "request": {
                        "path-param": {
                            "name": "param",
                            "value": "1"
                        }
                    },
                    "response": {
                        "content-type": "application/json",
                        "body": "{\"car\": \"1\"}",
                        "code": 200
                    }
                },{
                    "request": {
                        "path-param": {
                            "name": "param",
                            "value": "2"
                        }
                    },
                    "response": {
                        "content-type": "application/json",
                        "body": "{\"car\": \"2\"}",
                        "code": 200
                    }
                }
            ]
        }]
    }]
}
```
We can configure a query parameters and the response depends of the quey values
```json
{
    "port": 9001,
    "resources": [{
        "path": "/accounts/search",
        "methods": [{
                "name": "Recurso teste GET com query param",
				"type": "GET",
				"conversations": [{
					"request": {
						"query-param": [{
							"name": "num",
							"value": ["1"]
						}]
					},
					"response": {
                        "content-type": "application/json",
                        "body": "{\"conta\": \"1\"}",
                        "code": 200
                    }
				},{
					"request": {
						"query-param": [{
							"name": "num",
							"value": ["2"]
						}]
					},
					"response": {
						"content-type": "application/json",
						"body": "{\"conta\": \"2\"}",
						"code": 200
					}
				},{
					"request": {
						"query-param": [{
							"name": "num",
							"value": ["2"]
						},{
							"name": "plus",
							"value": ["2"]
						}]
					},
					"response": {
						"content-type": "application/json",
						"body": "{\"conta\": \"4\"}",
						"code": 200
					}
				}]
            }]
		}
    ]
}
```