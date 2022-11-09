package resolver

import (
	"github.com/kerti/idcra-api/model"
	"github.com/kerti/idcra-api/service"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
)

func (r *Resolver) CreateUser(ctx context.Context, args *struct {
	Email    string
	Password string
}) (*userResolver, error) {
	user := &model.User{
		Email:     args.Email,
		Password:  args.Password,
		IPAddress: *ctx.Value("requester_ip").(*string),
	}

	user, err := ctx.Value("userService").(*service.UserService).CreateUser(user)
	if err != nil {
		ctx.Value("log").(*logging.Logger).Errorf("Graphql error : %v", err)
		return nil, err
	}
	ctx.Value("log").(*logging.Logger).Debugf("Created user : %v", *user)
	return &userResolver{user}, nil
}

func (r *Resolver) ParentHasStudent(ctx context.Context, args *struct {
	UserId    string
	StudentId string
}) (*userResolver, error) {
	userStudent := &model.UsersStudentsRelations{
		UserId:    args.UserId,
		StudentId: args.StudentId,
	}

	user, err := ctx.Value("userService").(*service.UserService).CreateUserStudentRelation(userStudent)
	if err != nil {
		ctx.Value("log").(*logging.Logger).Errorf("Graphql error : %v", err)
		return nil, err
	}
	ctx.Value("log").(*logging.Logger).Debugf("Created user : %v", *user)
	return &userResolver{user}, nil
}

func (r *Resolver) RemoveStudentFromParent(ctx context.Context, args *struct {
	UserId    string
	StudentId string
}) (*userResolver, error) {
	userStudent := &model.UsersStudentsRelations{
		UserId:    args.UserId,
		StudentId: args.StudentId,
	}

	user, err := ctx.Value("userService").(*service.UserService).DeleteStudentFromParent(userStudent)
	if err != nil {
		ctx.Value("log").(*logging.Logger).Errorf("Graphql error : %v", err)
		return nil, err
	}
	ctx.Value("log").(*logging.Logger).Debugf("Created user : %v", *user)
	return &userResolver{user}, nil
}
