// package storage

// import (
// 	"fmt"
// 	"microblog/internal/models"
// 	"sync"
// )

// type ObjectStorage struct {
// 	mu         sync.Mutex
// 	usersOS    map[int]models.User
// 	postsOS    map[int]models.Post
// 	nextUserID int
// 	nextPostID int
// }

// func NewObjectStorage() *ObjectStorage {
// 	return &ObjectStorage{
// 		usersOS:    make(map[int]models.User),
// 		postsOS:    make(map[int]models.Post),
// 		nextUserID: 1,
// 		nextPostID: 1,
// 	}
// }

// func (o *ObjectStorage) CreateUser(user models.User) error {
// 	o.mu.Lock()
// 	defer o.mu.Unlock()
// 	user.ID = o.nextUserID
// 	o.nextUserID++
// 	o.usersOS[user.ID] = user
// 	return nil
// }

// func (o *ObjectStorage) CreatePost(post models.Post) (*models.Post, error) {
// 	o.mu.Lock()
// 	defer o.mu.Unlock()
// 	post.ID = o.nextPostID
// 	o.nextPostID++
// 	o.postsOS[post.ID] = post
// 	return &post, nil
// }

// func (o *ObjectStorage) GetUserByEmail(email string) (*models.User, error) {
// 	o.mu.Lock()
// 	defer o.mu.Unlock()
// 	for _, user := range o.usersOS {
// 		if user.Email == email {
// 			return &user, nil
// 		}
// 	}
// 	return nil, nil
// }

// func (o *ObjectStorage) GetPosts() ([]models.Post, error) {
// 	o.mu.Lock()
// 	defer o.mu.Unlock()
// 	posts := make([]models.Post, 0, len(o.postsOS))
// 	for _, post := range o.postsOS {
// 		posts = append(posts, post)
// 	}
// 	return posts, nil
// }

// func (o *ObjectStorage) GetPostById(id int) (*models.Post, error) {
// 	o.mu.Lock()
// 	defer o.mu.Unlock()
// 	post, exists := o.postsOS[id]
// 	if !exists {
// 		return nil, fmt.Errorf("post with id %d not found", id)
// 	}

// 	return &post, nil
// }

// func (o *ObjectStorage) LikePost(postID int) error {
// 	o.mu.Lock()
// 	defer o.mu.Unlock()

// 	post, exists := o.postsOS[postID]
// 	if !exists {
// 		return fmt.Errorf("post not found")
// 	}
// 	post.LikeCount++
// 	o.postsOS[postID] = post
// 	return nil
// }