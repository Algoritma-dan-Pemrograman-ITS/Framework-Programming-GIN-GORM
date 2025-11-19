# Authentication and Database Relationship Guide

This guide provides instructions on setting up user authentication and establishing relationships between users and blog posts in a Go web application using the Gin framework and GORM for database interactions.

## Authentication Setup

### Create new Controller: `AuthController.go`

1. Create a new file named `AuthController.go` in the `controllers` directory.
2. Create simple functions for user register (signup) and login.

   ### Login Function

   Login flow:

   - Accepts email and password.
   - Verifies user credentials.
   - Return JWT token on successful authentication.

   ```go
    func Login(c *gin.Context) {
        var body struct {
            Email    string `json:"email " binding:"required,email"`
            Password string `json:"password" binding:"required"`
        }

        if err := c.BindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
            return
        }

        // do database query to find user by email
        var user models.User
        if result := initializers.DB.First(&user, "email = ?", body.Email); result.Error != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
            return
        }

        // use bcrypt to compare password
        if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
            return
        }

        // create jwt
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "sub": strconv.FormatUint(uint64(user.ID), 10),
            // 5 seconds
            "exp": time.Now().Add(5 * time.Second).Unix(),
            // "exp": time.Now().Add(5 * time.Hour).Unix(), // 5 hours
            // "exp": time.Now().Add(5 * time.Minute).Unix(), // 5 minutes
        })

        secret := os.Getenv("JWT_SECRET")

        tokenString, err := token.SignedString([]byte(secret))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
            return
        }
        // set cookie named "token" with HttpOnly and a 24 hour expiryd
        maxAge := 24 * 60 * 60 // seconds

        // 5 seconds
        // maxAge := 5
        secure := os.Getenv("ENV") == "production"
        c.SetCookie("token", tokenString, maxAge, "/", "", secure, true)

        // return token in body for backward compatibility
        c.JSON(http.StatusOK, gin.H{"token": tokenString})
    }
   ```

   Basically, the login function verifies user credentials and generates a JWT token upon successful authentication. The JWT token can be sent to the client and stored in a cookie for subsequent requests or used in the Authorization header for every request that need authentication access.

   ### Register (Signup) Function

   Signup flow:

   - Accepts name, email, and password.
   - Hashes the password.
   - Stores user in the database.
   - Returns success message or error.

   ```go
    func Register(c *gin.Context) {

        // add binding tags for validation then use c.BindJSON
        var body struct {
            Name     string `json:"name" binding:"required,min=2"`
            Email    string `json:"email" binding:"required,email"`
            Password string `json:"password" binding:"required"`
        }

        if err := c.BindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // hash password
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
            return
        }


        user := models.User{Name: body.Name, Email: body.Email, Password: string(hashedPassword)}

        // insert user to database
        if result := initializers.DB.Create(&user); result.Error != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
            return
        }

        // avoid returning password
        user.Password = ""

        // or you can create new response struct without password field
        // userResponse := struct {
        //  ID    uint   `json:"id"`
        //  Name  string `json:"name"`
        //  Email string `json:"email"`
        // }{
        //  ID:    user.ID,
        //  Name:  user.Name,
        //  Email: user.Email,
        // }

        c.JSON(http.StatusOK, gin.H{"user": user})
        // c.JSON(http.StatusOK, gin.H{"user": userResponse})
    }
   ```

   Functionally, the register function handles user registration by validating input, hashing the password, and storing the new user in the database. It returns the created user object.

3. Create Login and Register routes in `main.go`:

   ```go
    router.POST("/signup", controllers.Register)
    router.POST("/signin", controllers.Login)
   ```

4. Now try to register and login using Postman or any API testing tool.

## Middleware for Authentication

Suppose we want to protect certain routes so that only authenticated users can access them. We can create an authentication middleware. Middleware basically intercepts requests before they reach the route handlers.

1. Create a new file named `AuthMiddleware.go` in the `middleware` directory.
2. Create the authentication middleware function:

   ```go
   func AuthMiddleware(c *gin.Context) {
        // Prefer token in cookie named "token". If missing, fall back to Authorization header.
        tokenString, err := c.Cookie("token")
        if err != nil || tokenString == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization cookie required"})
            return
        }

        //using bearer token from header u can use this also by uncommenting below code
        // authHeader := c.GetHeader("Authorization")
        // if authHeader == "" {
        //  c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
        //  return
        // }

        // if !strings.Contains(authHeader, "Bearer ") {
        //  c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
        //  return
        // }

        // //split the bearer and token
        // tokenString := strings.Split(authHeader, "Bearer ")[1]
        // if tokenString == "" {
        //  c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
        //  return
        // }

        // get jwt secret from env this is used to verify the token
        secret := os.Getenv("JWT_SECRET")

        // parse and verify token
        token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return []byte(secret), nil
        })

        // check for parsing errors or invalid token
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }

        sub, ok := claims["sub"]
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token subject"})
            return
        }

        // sub is numeric (float64) or string; handle both
        var id uint64
        parsed, err := strconv.ParseUint(sub.(string), 10, 64)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token subject"})
            return
        }
        id = parsed

        var user models.User
        if result := initializers.DB.First(&user, id); result.Error != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
            return
        }

        // set user in context for use in handlers so that we can access the authenticated user
        c.Set("user", user)

        // proceed to next handler
        c.Next()
   }

   ```

   This middleware firstly check for the JWT token in the cookie named "token". If the token is missing or invalid, it aborts the request with a 401 Unauthorized status. It then parses the token, retrieves the user ID from the token claims, and fetches the corresponding user from the database. If the user is found, it sets the user in the context for use in subsequent handlers.

3. Create function to get UserProfile that will use the authenticated user from context. Create this function in `AuthController.go`:

   ```go
   func Profile(c *gin.Context) {
        u, exists := c.Get("user")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

        // log.Printf("Context info: %v", u)

        user, ok := u.(models.User)
        if !ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read user from context"})
            return
        }

        user.Password = ""
        c.JSON(http.StatusOK, gin.H{"user": user})
    }
   ```

   This function retrieves the authenticated user from the context that was set by the `AuthMiddleware`. It checks if the user exists in the context, and if so, it returns the user information in the response. In real applications, you might want to just use user ID to fetch fresh data from the database instead of relying on the context data.

4. Create Profile route in `main.go`, and apply the `AuthMiddleware` to protect the route:

   ```go
    protected := router.Group("/")
    protected.Use(middleware.AuthMiddleware)
    {
        protected.GET("/profile", controllers.Profile)
    }
    // or u can do like this also (dont change the parameter order)
    // router.GET("/profile", middleware.AuthMiddleware, controllers.Profile)
   ```

   So when we defining middleware in the route, it will ensure that only authenticated users can access the `/profile` endpoint. The `AuthMiddleware` will run before the `Profile` handler, checking for a valid JWT token and setting the user in the context if authentication is successful (make sure to set the parameter correctly). So if you try to access `/profile` without a valid token, you will get a 401 Unauthorized response.

5. Now reload the server and try to access the `/profile` route with and without a valid JWT token to see the authentication in action.

## Database Relationship: User and BlogPost

### One-to-Many Relationship

In this section, we will establish a one-to-many relationship between `User` and `BlogPost` models. A user can have multiple blog posts, but each blog post belongs to a single user.

1. Update the `User` model to include a slice of `BlogPost`:

   ```go
   type User struct {
       gorm.Model
       Name     string     `gorm:"not null"`
       Email    string     `gorm:"uniqueIndex;not null"`
       Password string     `gorm:"not null"`
       Posts    []BlogPost `gorm:"foreignKey:UserID"` // One-to-Many relationship
   }
   ```

2. Update the `BlogPost` model to include a foreign key reference to the `User` model:

   ```go
    type BlogPost struct {
         gorm.Model
         Title   string `gorm:"not null"`
         Content string `gorm:"type:text;not null"`
         UserID  uint   // Foreign key to User
         User   User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // Belongs to User
    }
   ```

3. Update the database by reloading the server. GORM will automatically create the necessary foreign key constraints based on the model definitions.

4. Now in the creation of Blog, you must insert user_id also to associate the blog post with a user. For example, in your Blog creation controller, you can do something like this:

   ```go
    func CreateBlog(c *gin.Context) {
        var body struct {
            Title   string `json:"title" binding:"required"`
            Content string `json:"content" binding:"required"`
        }

        if err := c.BindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Get the authenticated user from context
        u, exists := c.Get("user")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

        user, ok := u.(models.User)
        if !ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read user from context"})
            return
        }

        blog := models.Blog{
            Title:   body.Title,
            Content: body.Content,
            UserID:  user.ID, // Associate blog post with the authenticated user
        }

        if result := initializers.DB.Create(&blog); result.Error != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"blog": blog})
    }
   ```

   Not quite different from before, but now we are associating the blog post with the authenticated user by setting the `UserID` field to the ID of the user retrieved from the context.

5. Now you can load blog posts along with their associated user information using GORM's `Preload` method. For example, when fetching blog posts, you can do:

   ```go
    var blogs []models.Blog
    initializers.DB.Preload("User").Find(&blogs)
   ```

   Using non‑eager loading you first load the blogs, then fetch each blog's user in a separate query (N+1 queries). Example:

   ```go
   // One query to load all blogs
   var blogs []models.Blog
   initializers.DB.Find(&blogs)

   // Then one query per blog to load its user
   for i := range blogs {
       var user models.User
       initializers.DB.First(&user, blogs[i].UserID)
       blogs[i].User = user
   }
   ```

   This executes SQL like:

   - SELECT \* FROM blog_posts;
   - For each blog: SELECT \* FROM users WHERE id = ? LIMIT 1;

   Which can be inefficient compared to using Preload (single JOIN/extra SELECT).

   Implementation in the Getter of Blogs controller:

   ```go
    func BlogGetAll(c *gin.Context) {
        //get all blogs
        var blogs []models.Blog
        result := initializers.DB.Preload("User").Find(&blogs)

        if result.Error != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": result.Error.Error(),
            })
            return
        }

        //return them
        c.JSON(http.StatusOK, gin.H{
            "blogs": blogs,
        })
   }
   ```

### Many-to-Many Relationship

In this example, we will make new Entity or Model named Class to simulate many-to-many relationship between User and Class. A user can enroll in multiple classes, and a class can have multiple users enrolled.

1. Create a new model named `Class`:

   ```go
   type Class struct {
        ID        uint   `gorm:"primaryKey" json:"id"`
        ClassName string `gorm:"size:255" json:"class_name"`
        ClassCode string `gorm:"size:100;unique" json:"class_code"`
        Users      []*User `gorm:"many2many:user_classes;" json:"users"`
    }
   ```

2. Update the `User` model to include a many-to-many relationship with `Class`:

   ```go
   type User struct {
        ID       uint   `gorm:"primaryKey" json:"id"`
        Name     string `gorm:"size:255" json:"name"`
        Email    string `gorm:"unique" json:"email"`
        Password string `gorm:"size:255" json:"-"`
        // add this line to establish relationship with Blog model
        Blogs   []Blog   `gorm:"foreignKey:UserID" json:"blogs"`
        Classes []*Class `gorm:"many2many:user_classes;" json:"classes"`
    }
   ```

3. Update the database by reloading the server. GORM will automatically create the necessary join table (`user_classes`) based on the model definitions.

4. Create controller functions to handle class creation and fetching participants:

   ```go
    func CreateClass(c *gin.Context) {
        var body struct {
            ClassName string `json:"class_name" binding:"required"`
            ClassCode string `json:"class_code" binding:"required"`
            UserIDs   []uint `json:"user_ids"` // IDs of users to enroll in the class
        }

        if err := c.BindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        class := models.Class{
            ClassName: body.ClassName,
            ClassCode: body.ClassCode,
        }

        // fetch users based on provided IDs
        var users []*models.User
        for _, id := range body.UserIDs {
            var user models.User
            if result := initializers.DB.First(&user, id); result.Error != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "User not found with ID: " + strconv.Itoa(int(id))})
                return
            }
            users = append(users, &user)
        }

        class.Users = users

        if result := initializers.DB.Create(&class); result.Error != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"class": class})
    }
   ```

5. Create route in `main.go` for class creation and fetching participants:

   ```go
    router.POST("/create-class", controllers.CreateClass)
   ```

6. Now try to create classes and enroll users using Postman or any API testing tool. Check the database to see the relationships established in the `user_classes` join table.

7. Get Participants of a Class & user enrolled classes:

   ```go
    func GetClassParticipants(c *gin.Context) {
        classID := c.Param("id")
        var class models.Class
        if result := initializers.DB.Preload("Users").First(&class, classID); result.Error != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Class not found"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"users": class.Users})
    }

    func GetClassesUserEnrolled(c *gin.Context) {
        userID := c.Param("id")
        var user models.User
        if result := initializers.DB.Preload("Classes").First(&user, userID); result.Error != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"classes": user.Classes})
    }
   ```

8. Create route in `main.go` for fetching class participants:

   ```go
    router.GET("/class/:id/participants", controllers.GetClassParticipants)
    router.GET("/user/:id/classes", controllers.GetClassesUserEnrolled)
   ```

9. Now try to get class participants and user enrolled classes using Postman or any API testing tool.

### Updating Many-to-Many Relationships

1. Create new function in controller that will update the users enrolled in a class:

   ```go
   func UpdateClassParticipants(c *gin.Context) {
        classID := c.Param("id")
        var body struct {
            UserIDs []uint `json:"user_ids" binding:"required"`
        }
        if err := c.BindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var class models.Class
        if result := initializers.DB.First(&class, classID); result.Error != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Class not found"})
            return
        }

        // fetch users based on provided IDs
        var users []*models.User
        for _, id := range body.UserIDs {
            var user models.User
            if result := initializers.DB.First(&user, id); result.Error != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "User not found with ID: " + strconv.Itoa(int(id))})
                return
            }
            users = append(users, &user)
        }

        // use GORM association to replace many2many relations
        if err := initializers.DB.Model(&class).Association("Users").Replace(users); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // reload class with users to return updated relation
        if result := initializers.DB.Preload("Users").First(&class, classID); result.Error != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"class": class})
   }

   ```

   In this function, we first retrieve the class by its ID using GIN Parameter. We then parse the request body to get the list of user IDs that should be enrolled in the class. Then we check if the class and each user exists in the database. Finally, we use GORM's `Association("Users").Replace(users)` method to update the many-to-many relationship by replacing the existing users with the new list. After updating, we reload the class with its associated users to return the updated relationship in the response.

2. Create route in `main.go` for updating class participants:

   ```go
   router.PUT("/class/:id/participants", controllers.UpdateClassParticipants)
   ```

### Deleting Records from Many-to-Many Relationships

When working with many-to-many relationships, you often need to remove specific associations between entities. GORM provides convenient methods to handle these operations through its Association API.

#### Deleting a Single User from a Class

To remove a specific user from a class, you can use the `Delete` method on the association:

```go
func DeleteParticipantFromClass(c *gin.Context) {
    classID := c.Param("id")
    userID := c.Param("user_id")

    var class models.Class
    if result := initializers.DB.First(&class, classID); result.Error != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Class not found"})
        return
    }

    var user models.User
    if result := initializers.DB.First(&user, userID); result.Error != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
    return
    }

    // use GORM association to delete many2many relation
    if err := initializers.DB.Model(&class).Association("Users").Delete(&user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User removed from class successfully"})
}
```

This function:

- Retrieves both the class and user from the database using their IDs from URL parameters
- Uses `Association("Users").Delete(&user)` to remove the relationship between the specific user and class
- Only deletes the relationship record in the `user_classes` join table, not the actual user or class records
- Returns a success message confirming the removal

#### Deleting a User from All Classes

To remove a user from all classes they're enrolled in, use the `Clear` method:

```go
func DeleteUserFromAllClasses(c *gin.Context) {
    userID := c.Param("id")

    var user models.User
    if result := initializers.DB.First(&user, userID); result.Error != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
    return
    }

    // use GORM association to clear many2many relations
    if err := initializers.DB.Model(&user).Association("Classes").Clear(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
    }

c.JSON(http.StatusOK, gin.H{"message": "User removed from all classes successfully"})
}
```

This function:

- Retrieves the user from the database using the ID from URL parameter
- Uses `Association("Classes").Clear()` to remove all relationships between the user and their classes
- Deletes all entries in the `user_classes` join table where the user is involved
- Does not delete the user record itself or any class records

#### Key Differences Between Delete and Clear

- **Delete**: Removes specific association(s). You pass the related record(s) you want to unlink.
- **Clear**: Removes all associations for that model. No parameters needed - it clears everything.

#### Create Routes for Deletion Operations

Add these routes in `main.go`:

```go
router.DELETE("/class/:id/participants/:user_id", controllers.DeleteParticipantFromClass)
router.DELETE("/user/:id/classes", controllers.DeleteUserFromAllClasses)
```

Now you can test these deletion endpoints:

- `DELETE /class/1/participants/5` - Removes user with ID 5 from class with ID 1
- `DELETE /user/5/classes` - Removes user with ID 5 from all classes they're enrolled in
