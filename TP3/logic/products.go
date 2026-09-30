package logic

import "errors"

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func ValidateProduct(product Product) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.Price <= 0 {
		return errors.New("product price must be greater than 0")
	}
	return nil
}

func ApllyDiscount(product Product, discount float64) Product {
	product.Price = product.Price * (1 - discount/100)
	return product
}
