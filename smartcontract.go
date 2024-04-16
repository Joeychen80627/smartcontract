package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type SupplyChainContract struct {
	contractapi.Contract
}

func (s *SupplyChainContract) getTimestamp(ctx contractapi.TransactionContextInterface) (string, error) {
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	return time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339), nil
}

func (s *SupplyChainContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	products := []Product{
		{ID: "p1", Name: "Laptop", Status: "Manufactured", Owner: "CompanyA", CreatedAt: timestamp, UpdatedAt: timestamp, Description: "High-end gaming laptop", Category: "Electronics"},
		{ID: "p2", Name: "Smartphone", Status: "Manufactured", Owner: "CompanyB", CreatedAt: timestamp, UpdatedAt: timestamp, Description: "Latest model smartphone", Category: "Electronics"},
	}

	for _, product := range products {
		if err := s.putProduct(ctx, &product); err != nil {
			return err
		}
	}

	return nil
}

func (s *SupplyChainContract) CreateProduct(ctx contractapi.TransactionContextInterface, id, name, owner, description, category string) error {
	exists, err := s.ProductExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("product with ID %s already exists", id)
	}

	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	newProduct := Product{
		ID:          id,
		Name:        name,
		Status:      "Manufactured",
		Owner:       owner,
		CreatedAt:   timestamp,
		UpdatedAt:   timestamp,
		Description: description,
		Category:    category,
	}

	err = s.putProduct(ctx, &newProduct)
	if err != nil {
		return fmt.Errorf("failed to put product into ledger: %v", err)
	}

	return nil
}

func (s *SupplyChainContract) UpdateProduct(ctx contractapi.TransactionContextInterface, id string, newStatus string, newOwner string, newDescription string, newCategory string) error {
	exists, err := s.ProductExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("product with ID %s does not exist", id)
	}

	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	existingProduct, err := s.QueryProduct(ctx, id)
	if err != nil {
		return err
	}

	if existingProduct.Owner != newOwner {
		existingProduct.Owner = newOwner
		existingProduct.UpdatedAt = timestamp
	}

	existingProduct.Status = newStatus
	existingProduct.Description = newDescription
	existingProduct.Category = newCategory
	existingProduct.UpdatedAt = timestamp

	err = s.putProduct(ctx, existingProduct)
	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	return nil
}

func (s *SupplyChainContract) TransferOwnership(ctx contractapi.TransactionContextInterface, id, newOwner string) error {
	exists, err := s.ProductExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("product with ID %s does not exist", id)
	}

	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	existingProduct, err := s.QueryProduct(ctx, id)
	if err != nil {
		return err
	}

	existingProduct.Owner = newOwner
	existingProduct.UpdatedAt = timestamp

	err = s.putProduct(ctx, existingProduct)
	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	return nil
}

func (s *SupplyChainContract) QueryProduct(ctx contractapi.TransactionContextInterface, id string) (*Product, error) {
	exists, err := s.ProductExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("the product with ID %s does not exist", id)
	}

	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read product from ledger: %v", err)
	}
	if productJSON == nil {
		return nil, fmt.Errorf("the product with ID %s does not exist", id)
	}

	var product Product
	err = json.Unmarshal(productJSON, &product)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal product JSON: %v", err)
	}

	return &product, nil
}

func (s *SupplyChainContract) putProduct(ctx contractapi.TransactionContextInterface, product *Product) error {
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(product.ID, productJSON)
}

func (s *SupplyChainContract) ProductExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return productJSON != nil, nil
}

func (s *SupplyChainContract) GetAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var products []*Product
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var product Product
		if err := json.Unmarshal(queryResponse.Value, &product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SupplyChainContract{})
	if err != nil {
		fmt.Printf("Error creating supply chain chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting supply chain chaincode: %s", err.Error())
	}
}

}
