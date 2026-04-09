package router

const swaggerSpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "SomeWebProject API",
    "version": "1.0.0",
    "description": "Authentication, users and products API"
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "Local development server"
    }
  ],
  "tags": [
    {"name": "Auth", "description": "Authentication and token management"},
    {"name": "Users", "description": "User administration"},
    {"name": "Products", "description": "Product catalog management"}
  ],
  "paths": {
    "/api/auth/register": {
      "post": {
        "tags": ["Auth"],
        "summary": "Register a new user",
        "operationId": "registerUser",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {"$ref": "#/components/schemas/RegisterRequest"},
              "examples": {
                "default": {
                  "value": {"email": "user@example.com", "password": "secret123", "gender": "male", "age": 25}
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "User created",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/UserResponse"}
              }
            }
          },
          "400": {"$ref": "#/components/responses/BadRequest"}
        }
      }
    },
    "/api/auth/login": {
      "post": {
        "tags": ["Auth"],
        "summary": "Login and receive JWT tokens",
        "operationId": "loginUser",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {"$ref": "#/components/schemas/LoginRequest"},
              "examples": {
                "default": {
                  "value": {"email": "user@example.com", "password": "secret123"}
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Authenticated",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/AuthResponse"}
              }
            }
          },
          "401": {"$ref": "#/components/responses/Unauthorized"}
        }
      }
    },
    "/api/auth/refresh": {
      "post": {
        "tags": ["Auth"],
        "summary": "Refresh access and refresh tokens",
        "operationId": "refreshTokens",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {"$ref": "#/components/schemas/RefreshRequest"}
            }
          }
        },
        "responses": {
          "200": {
            "description": "Tokens refreshed",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/AuthResponse"}
              }
            }
          },
          "401": {"$ref": "#/components/responses/Unauthorized"}
        }
      }
    },
    "/api/auth/me": {
      "get": {
        "tags": ["Auth"],
        "summary": "Get current authenticated user",
        "security": [{"bearerAuth": []}],
        "responses": {
          "200": {
            "description": "Current user",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/UserResponse"}
              }
            }
          },
          "401": {"$ref": "#/components/responses/Unauthorized"}
        }
      }
    },
    "/api/users": {
      "get": {
        "tags": ["Users"],
        "summary": "List all users",
        "security": [{"bearerAuth": []}],
        "responses": {
          "200": {
            "description": "User list",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {"$ref": "#/components/schemas/UserResponse"}
                }
              }
            }
          },
          "401": {"$ref": "#/components/responses/Unauthorized"},
          "403": {"$ref": "#/components/responses/Forbidden"}
        }
      }
    },
    "/api/users/{id}": {
      "parameters": [
        {"$ref": "#/components/parameters/UserID"}
      ],
      "get": {
        "tags": ["Users"],
        "summary": "Get user by ID",
        "security": [{"bearerAuth": []}],
        "responses": {
          "200": {
            "description": "User found",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/UserResponse"}
              }
            }
          },
          "401": {"$ref": "#/components/responses/Unauthorized"},
          "403": {"$ref": "#/components/responses/Forbidden"},
          "404": {"$ref": "#/components/responses/NotFound"}
        }
      },
      "put": {
        "tags": ["Users"],
        "summary": "Update user information",
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {"$ref": "#/components/schemas/UpdateUserRequest"}
            }
          }
        },
        "responses": {
          "200": {
            "description": "User updated",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/UserResponse"}
              }
            }
          },
          "400": {"$ref": "#/components/responses/BadRequest"},
          "401": {"$ref": "#/components/responses/Unauthorized"},
          "403": {"$ref": "#/components/responses/Forbidden"},
          "404": {"$ref": "#/components/responses/NotFound"}
        }
      },
      "delete": {
        "tags": ["Users"],
        "summary": "Block user",
        "security": [{"bearerAuth": []}],
        "responses": {
          "204": {"description": "User blocked"},
          "401": {"$ref": "#/components/responses/Unauthorized"},
          "403": {"$ref": "#/components/responses/Forbidden"},
          "404": {"$ref": "#/components/responses/NotFound"}
        }
      }
    },
    "/api/products": {
      "get": {
        "tags": ["Products"],
        "summary": "List all products",
        "security": [{"bearerAuth": []}],
        "responses": {
          "200": {
            "description": "Product list",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {"$ref": "#/components/schemas/ProductResponse"}
                }
              }
            }
          },
          "401": {"$ref": "#/components/responses/Unauthorized"}
        }
      },
      "post": {
        "tags": ["Products"],
        "summary": "Create product",
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {"$ref": "#/components/schemas/ProductCreateRequest"}
            }
          }
        },
        "responses": {
          "201": {
            "description": "Product created",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ProductResponse"}
              }
            }
          },
          "400": {"$ref": "#/components/responses/BadRequest"},
          "401": {"$ref": "#/components/responses/Unauthorized"},
          "403": {"$ref": "#/components/responses/Forbidden"}
        }
      }
    },
    "/api/products/{id}": {
      "parameters": [
        {"$ref": "#/components/parameters/ProductID"}
      ],
      "get": {
        "tags": ["Products"],
        "summary": "Get product by ID",
        "security": [{"bearerAuth": []}],
        "responses": {
          "200": {
            "description": "Product found",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ProductResponse"}
              }
            }
          },
          "401": {"$ref": "#/components/responses/Unauthorized"},
          "404": {"$ref": "#/components/responses/NotFound"}
        }
      },
      "put": {
        "tags": ["Products"],
        "summary": "Update product",
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {"$ref": "#/components/schemas/ProductUpdateRequest"}
            }
          }
        },
        "responses": {
          "200": {
            "description": "Product updated",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/ProductResponse"}
              }
            }
          },
          "400": {"$ref": "#/components/responses/BadRequest"},
          "401": {"$ref": "#/components/responses/Unauthorized"},
          "403": {"$ref": "#/components/responses/Forbidden"},
          "404": {"$ref": "#/components/responses/NotFound"}
        }
      },
      "delete": {
        "tags": ["Products"],
        "summary": "Delete product",
        "security": [{"bearerAuth": []}],
        "responses": {
          "204": {"description": "Product deleted"},
          "401": {"$ref": "#/components/responses/Unauthorized"},
          "403": {"$ref": "#/components/responses/Forbidden"},
          "404": {"$ref": "#/components/responses/NotFound"}
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "bearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    },
    "parameters": {
      "UserID": {
        "name": "id",
        "in": "path",
        "required": true,
        "schema": {"type": "integer", "format": "int64", "minimum": 1},
        "description": "User identifier"
      },
      "ProductID": {
        "name": "id",
        "in": "path",
        "required": true,
        "schema": {"type": "integer", "format": "int64", "minimum": 1},
        "description": "Product identifier"
      }
    },
    "responses": {
      "BadRequest": {
        "description": "Invalid request",
        "content": {
          "application/json": {
            "schema": {"$ref": "#/components/schemas/ErrorResponse"}
          }
        }
      },
      "Unauthorized": {
        "description": "Unauthorized",
        "content": {
          "application/json": {
            "schema": {"$ref": "#/components/schemas/ErrorResponse"}
          }
        }
      },
      "Forbidden": {
        "description": "Forbidden",
        "content": {
          "application/json": {
            "schema": {"$ref": "#/components/schemas/ErrorResponse"}
          }
        }
      },
      "NotFound": {
        "description": "Not found",
        "content": {
          "application/json": {
            "schema": {"$ref": "#/components/schemas/ErrorResponse"}
          }
        }
      }
    },
    "schemas": {
      "RegisterRequest": {
        "type": "object",
        "required": ["email", "password", "gender", "age"],
        "properties": {
          "email": {"type": "string", "format": "email"},
          "password": {"type": "string", "format": "password"},
          "gender": {"type": "string"},
          "age": {"type": "integer", "format": "int32", "minimum": 0}
        }
      },
      "LoginRequest": {
        "type": "object",
        "required": ["email", "password"],
        "properties": {
          "email": {"type": "string", "format": "email"},
          "password": {"type": "string", "format": "password"}
        }
      },
      "RefreshRequest": {
        "type": "object",
        "required": ["refresh_token"],
        "properties": {
          "refresh_token": {"type": "string"}
        }
      },
      "UserResponse": {
        "type": "object",
        "properties": {
          "id": {"type": "integer", "format": "int64"},
          "email": {"type": "string", "format": "email"},
          "role": {"type": "string", "enum": ["user", "seller", "admin"]},
          "age": {"type": "integer", "format": "int32"},
          "gender": {"type": "string"},
          "is_blocked": {"type": "boolean"},
          "created_at": {"type": "string", "format": "date-time"},
          "updated_at": {"type": "string", "format": "date-time"}
        }
      },
      "AuthResponse": {
        "type": "object",
        "properties": {
          "access_token": {"type": "string"},
          "refresh_token": {"type": "string"},
          "user": {"$ref": "#/components/schemas/UserResponse"}
        }
      },
      "UpdateUserRequest": {
        "type": "object",
        "properties": {
          "email": {"type": "string", "format": "email"},
          "age": {"type": "integer", "format": "int32", "minimum": 0},
          "gender": {"type": "string"},
          "role": {"type": "string", "enum": ["user", "seller", "admin"]}
        }
      },
      "ProductCreateRequest": {
        "type": "object",
        "required": ["name", "description", "price", "stock"],
        "properties": {
          "name": {"type": "string"},
          "description": {"type": "string"},
          "price": {"type": "number", "format": "double", "minimum": 0},
          "stock": {"type": "integer", "format": "int32", "minimum": 0}
        }
      },
      "ProductUpdateRequest": {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "description": {"type": "string"},
          "price": {"type": "number", "format": "double", "minimum": 0},
          "stock": {"type": "integer", "format": "int32", "minimum": 0}
        }
      },
      "ProductResponse": {
        "type": "object",
        "properties": {
          "id": {"type": "integer", "format": "int64"},
          "name": {"type": "string"},
          "description": {"type": "string"},
          "price": {"type": "number", "format": "double"},
          "stock": {"type": "integer", "format": "int32"},
          "owner_id": {"type": "integer", "format": "int64"},
          "created_at": {"type": "string", "format": "date-time"},
          "updated_at": {"type": "string", "format": "date-time"}
        }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "error": {"type": "string"}
        }
      }
    }
  }
}`
