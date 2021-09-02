package main

import (
	"context"
	"errors"
	"log"
	"net"

	casinopb "github.com/preslavmihaylov/go-grpc-crash-course/gen/casino"
	commonpb "github.com/preslavmihaylov/go-grpc-crash-course/gen/common"
	"github.com/preslavmihaylov/go-grpc-crash-course/gen/payment_statements"
	"google.golang.org/grpc"
)

type userID string

var (
	tokensPerDollar         = int32(5)
	casinoAddr              = "localhost:10000"
	paymentStatementsAddr   = "localhost:10001"
	paymentStatementsClient payment_statements.PaymentStatementsClient
)

func main() {
	var conn *grpc.ClientConn
	paymentStatementsClient, conn = setupPaymentStatementsClient()
	defer conn.Close()

	log.Println("Successfully connected to payment_statements...")

	grpcServer, lis := setupCasinoServer()
	grpcServer.Serve(lis)
}

func setupCasinoServer() (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", casinoAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	casinopb.RegisterCasinoServer(grpcServer, newCasinoServer())

	return grpcServer, lis
}

func setupPaymentStatementsClient() (payment_statements.PaymentStatementsClient, *grpc.ClientConn) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(paymentStatementsAddr, opts...)

	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	return payment_statements.NewPaymentStatementsClient(conn), conn
}

func newCasinoServer() *casinoServer {
	return &casinoServer{
		stockPrice:     100,
		userToTokens:   map[userID]int32{},
		userToPayments: map[userID][]int32{},
		userToStocks:   map[userID]int32{},
	}
}

func (c *casinoServer) BuyTokens(ctx context.Context, payment *commonpb.Payment) (*casinopb.Tokens, error) {
	log.Printf("BuyTokens invoked with payment %v\n", payment)

	usrID := userID(payment.User.GetId())
	tokens := payment.GetAmount() * tokensPerDollar

	c.userToPayments[usrID] = append(c.userToPayments[usrID], -payment.Amount)
	c.userToTokens[usrID] += tokens

	return &casinopb.Tokens{Count: tokens}, nil
}

func (c *casinoServer) Withdraw(ctx context.Context, withdrawReq *casinopb.WithdrawRequest) (*commonpb.Payment, error) {
	toWithdraw := withdrawReq.GetTokensCnt()
	log.Printf("Withdraw invoked with tokens %v\n", toWithdraw)

	usrID := userID(withdrawReq.User.GetId())

	if !c.hasEnoughTokens(usrID, toWithdraw) {
		return nil, errors.New("not enough tokens to withdraw")
	}

	amount := toWithdraw / tokensPerDollar
	c.userToTokens[usrID] -= toWithdraw
	c.userToPayments[usrID] = append(c.userToPayments[usrID], amount)

	return &commonpb.Payment{
		User:   withdrawReq.User,
		Amount: amount,
	}, nil
}

func (c *casinoServer) GetTokenBalance(ctx context.Context, user *commonpb.User) (*casinopb.Tokens, error) {
	log.Printf("GetTokenBalance invoked with user %v\n", user)

	usrID := userID(user.GetId())
	return &casinopb.Tokens{Count: c.userToTokens[usrID]}, nil
}

func (c *casinoServer) GetPayments(user *commonpb.User, stream casinopb.Casino_GetPaymentsServer) error {
	panic("not implmented")
}

func (c *casinoServer) GetPaymentStatement(ctx context.Context, user *commonpb.User) (*commonpb.PaymentStatement, error) {
	panic("not implemented")
}

func (c *casinoServer) Gamble(stream casinopb.Casino_GambleServer) error {
	panic("not implemented")
}

type casinoServer struct {
	stockPrice int32

	userToTokens   map[userID]int32
	userToPayments map[userID][]int32
	userToStocks   map[userID]int32
}
