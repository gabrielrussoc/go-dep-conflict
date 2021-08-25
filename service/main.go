package main

import (
    "github.com/google/uuid"
    "databricks.com/common"
)

func main() {
    common.F(uuid.New())
}
