/**
 * BLOCKCHAIN EXPERIENCE
 * Based on Monopoly Game Rules.
 * 
 */

package main

import (
//	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

/**
 * Property corresponds to a property title.
 * The Monopoly's property title has a lot of attributes such as: name, building block, terrain price, rent price, building price and others.
 * At this version of Blockchain Experience only name, value and current holder are used.
 */
type Property struct {
	Name string `json:"Name"`
	Value int `json:"Value"`
	Holder string `json:"Holder"`
}

/**
 * Wallet contains the current cash balance of the Holder/Owner.
 * In the future it can contains the Properties and other assets stored as Tokens.
 */
type Wallet struct {
	Value int `json:"Value"`
	Holder string `json:"Holder"`
	Status string `json:"Status"`
}


/**
 * The Init method
 * called when the Smart Contract is instantiated by the network
 * When deployed in Blockchain it initializes all properties and wallets.
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	s.initGame(APIstub)
	return shim.Success(nil)
}


/**
 * The Query method
 * LEGACY method
 */
func (s *SmartContract) Query(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Error("Unknown supported call - Query()")
}



/**
 * The Invoke method
 * called when an application requests to run the Smart Contract 
 * The app also specifies the specific smart contract function to call with args
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryAllProperties" {
		return s.queryAllProperties(APIstub)
	} else if function == "queryProperty" {
		return s.queryProperty(APIstub, args)
	} else if function == "queryWallet" {
		return s.queryWallet(APIstub)
	} else if function == "queryAllWallets" {
		return s.queryAllWallets(APIstub)
	} else if function == "queryPropertyHistory" {
		return s.queryPropertyHistory(APIstub)
	} else if function == "queryWalletHistory" {
		return s.queryWalletHistory(APIstub)
	} else if function == "initGame" {
		return s.initGame(APIstub)
	} else if function == "transferProperty" {
		return s.transferProperty(APIstub, args)
	} else if function == "pay" {
		return s.pay(APIstub)
	} else if function == "bankrupt" {
		return s.bankrupt(APIstub)
	}

	return shim.Error("Invalid Smart Contract function name.")
}


/**
 * The queryAllProperties method
 *
 */
func (s *SmartContract) queryAllProperties(APIstub shim.ChaincodeStubInterface) sc.Response {
	var properties []Property

	initialPropertiesList := getInitialStateProperties()

	i := 0
	for i < len(initialPropertiesList) {
		
		propertyAsBytes, _ := APIstub.GetState(initialPropertiesList[i].Name)

		var property Property
		json.Unmarshal(propertyAsBytes, &property)
		properties = append(properties, property)
		
		i = i + 1
	}
	
	propertiesAsBytes, _ := json.Marshal(properties)

	return shim.Success(propertiesAsBytes)
}


/**
 * The queryProperty method
 *
 */
func (s *SmartContract) queryProperty(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	var arg, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	
	arg = args[0]
	propertyAsBytes, err := APIstub.GetState(arg)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + arg + "\"}"
		return shim.Error(jsonResp)
	} else if propertyAsBytes == nil {
		jsonResp = "{\"Error\":\"Property does not exist: " + arg + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(propertyAsBytes)
}


/**
 * The queryWallet method
 *
 */
func (s *SmartContract) queryWallet(APIstub shim.ChaincodeStubInterface) sc.Response {
	var wallet string
	var err error
	
	_, args := APIstub.GetFunctionAndParameters()
	
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}
	
	wallet = args[0]
	walletAsBytes, err := APIstub.GetState(wallet)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get wallet for %s", wallet))
	} else if walletAsBytes == nil {
		return shim.Error(fmt.Sprintf("Wallet does not exist: %s", wallet))
	}
	return shim.Success(walletAsBytes)
}


/**
 * The queryAllWallets method
 *
 */
func (s *SmartContract) queryAllWallets(APIstub shim.ChaincodeStubInterface) sc.Response {
	var wallets []Wallet
	var wallet Wallet

	// from Player 1 to Player 6
	i := 1
	for (i <= 6) {
		holder := fmt.Sprintf("Player %d", i)
		walletAsBytes, _ := APIstub.GetState(holder)
		json.Unmarshal(walletAsBytes, &wallet)
		wallets = append(wallets, wallet)
		
		i = i + 1
	}
	
	/** Add BANK wallet */
	walletAsBytes, _ := APIstub.GetState("Bank")
	json.Unmarshal(walletAsBytes, &wallet)
	wallets = append(wallets, wallet)
	
	// return
	walletsAsBytes, _ := json.Marshal(wallets)

	return shim.Success(walletsAsBytes)
}


/**
 * The queryPropertyHistory method.
 *
 */
func (s *SmartContract) queryPropertyHistory(APIstub shim.ChaincodeStubInterface) sc.Response {
	type PropertyHistory struct {
		TxId    string   `json:"txId"`
		Value   Property  `json:"value"`
	}

	var history []PropertyHistory;
	var property Property

	_, args := APIstub.GetFunctionAndParameters()
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}
	
	propertyName := args[0]
	fmt.Printf("- start getHistoryForProperty: %s\n", propertyName)
	
	// Get History
	resultsIterator, err := APIstub.GetHistoryForKey(propertyName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		historyData, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var tx PropertyHistory
		tx.TxId = historyData.TxId                     //copy transaction id over
		json.Unmarshal(historyData.Value, &property)     //un stringify it aka JSON.parse()
		if historyData.Value == nil {                  //marble has been deleted
			var emptyProperty Property
			tx.Value = emptyProperty 		//copy nil marble
		} else {
			json.Unmarshal(historyData.Value, &property) //un stringify it aka JSON.parse()
			tx.Value = property                      //copy marble over
		}
		history = append(history, tx)              //add this tx to the list
	}

	historyAsBytes, _ := json.Marshal(history)     //convert to array of bytes

	return shim.Success(historyAsBytes)
}

/**
 * The queryWalletHistory method
 *
 */
func (s *SmartContract) queryWalletHistory(APIstub shim.ChaincodeStubInterface) sc.Response {
	type WalletHistory struct {
		TxId    string   `json:"txId"`
		Value   Wallet  `json:"value"`
	}

	var history []WalletHistory;
	var wallet Wallet

	_, args := APIstub.GetFunctionAndParameters()
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}
	
	walletName := args[0]
		
	// Get History
	resultsIterator, err := APIstub.GetHistoryForKey(walletName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		historyData, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var tx WalletHistory
		tx.TxId = historyData.TxId                     //copy transaction id over
		json.Unmarshal(historyData.Value, &wallet)     //un stringify it aka JSON.parse()
		if historyData.Value == nil {                  //marble has been deleted
			var emptyWallet Wallet
			tx.Value = emptyWallet 		//copy nil marble
		} else {
			json.Unmarshal(historyData.Value, &wallet) //un stringify it aka JSON.parse()
			tx.Value = wallet                      //copy marble over
		}
		history = append(history, tx)              //add this tx to the list
	}

	historyAsBytes, _ := json.Marshal(history)     //convert to array of bytes

	return shim.Success(historyAsBytes)
}




/**
 * The initGame method
 * 
 */
func (s *SmartContract) initGame(APIstub shim.ChaincodeStubInterface) sc.Response {
	s.initProperties(APIstub)
	s.initWallets(APIstub)
	return shim.Success(nil)
}


/**
 * The finishGame method
 * This method was not implement yet.
 */
func (s *SmartContract) finishGame(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}


/**
 * The getInitialStateProperties function
 *
 */
func getInitialStateProperties() []Property {
	return []Property{
		Property{Name:"Ipanema", Value:220, Holder:"Bank" },
		Property{Name:"Leblon", Value:220, Holder:"Bank" },
		Property{Name:"Copacabana", Value:240, Holder:"Bank" },
		Property{Name:"Avenida Brigadeiro Faria Lima", Value:200, Holder:"Bank" },
		Property{Name:"Avenida Presidente Juscelino Kubistcheck", Value:180, Holder:"Bank" },
		Property{Name:"Avenida Engenheiro Luis Carlos Berrini", Value:180, Holder:"Bank" },
		Property{Name:"Avenida Atlantica", Value:160, Holder:"Bank" },
		Property{Name:"Avenida Vieira Souto", Value:140, Holder:"Bank" },
		Property{Name:"Niteroi", Value:140, Holder:"Bank" },
		Property{Name:"Avenida Paulista", Value:120, Holder:"Bank" },
		Property{Name:"Rua 25 de Marco", Value:100, Holder:"Bank" },
		Property{Name:"Avenida Sao Joao", Value:100, Holder:"Bank" },
		Property{Name:"Praca da Se", Value:60, Holder:"Bank" },
		Property{Name:"Avenida Sumare", Value:60, Holder:"Bank" },
		Property{Name:"Avenida Cidade Jardim", Value:260, Holder:"Bank" },
		Property{Name:"Pacaembu", Value:260, Holder:"Bank" },
		Property{Name:"Ibirapuera", Value:280, Holder:"Bank" },
		Property{Name:"Barra da Tijuca", Value:300, Holder:"Bank" },
		Property{Name:"Jardim Botanico", Value:300, Holder:"Bank" },
		Property{Name:"Lagoa Rodrigo de Freitas", Value:320, Holder:"Bank" },
		Property{Name:"Avenida Morumbi", Value:350, Holder:"Bank" },
		Property{Name:"Rua Oscar Freire", Value:400, Holder:"Bank" },
	}
}


/**
 * The initProperties method
 *
 */
func (s *SmartContract) initProperties(APIstub shim.ChaincodeStubInterface) sc.Response {
	properties := getInitialStateProperties()
	i := 0
	for i < len(properties) {
		fmt.Println("i is ", i)
		propertyAsBytes, _ := json.Marshal(properties[i])
		nameAsBytes, _ := json.Marshal(properties[i].Name)
		APIstub.PutState(fmt.Sprintf("Property %i", i), nameAsBytes)
		APIstub.PutState(properties[i].Name, propertyAsBytes)
		fmt.Println("Added", properties[i])
		i = i + 1
	}
	return shim.Success(nil)
}


/**
 * The initWallets method
 * Start six wallets with names Player 1 to Player 6 with initial balance 1,500.
 * Add an additional wallet named Bank with initial value 10,000.
 */
func (s *SmartContract) initWallets(APIstub shim.ChaincodeStubInterface) sc.Response {
        i := 1
	for (i <= 6) {
		fmt.Println("i is ", i)
		currentHolder := fmt.Sprintf("Player %d", i)
		wallet := Wallet{Holder: currentHolder, Value: 1500, Status: "Active"}

		walletAsBytes, _ := json.Marshal(wallet)
		APIstub.PutState(wallet.Holder, walletAsBytes)
		fmt.Println("Added", wallet)
		i = i + 1
	}
	
	bankWallet := Wallet{Holder: "Bank", Value: 100000, Status: "Active"}
	bankWalletAsBytes, _ := json.Marshal(bankWallet)
	APIstub.PutState("Bank", bankWalletAsBytes)
	
	return shim.Success(nil)
}


/**
 * The transferProperty method
 * transferProperty( property, currentHolder, newHolder, price)
 * property is the Property Name, e.g. Ipanema
 * currentHolder is the seller, e.g. Player 1
 * newHolder is the buyer, e.g. Player 2
 * price is the price, e.g. 200
 */
func (s *SmartContract) transferProperty(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	
	argProperty := args[0]
	argCurrentHolder := args[1]
	argNewHolder := args[2]
	argPrice := args[3]
		
	propertyAsBytes, _ := APIstub.GetState(argProperty)
	if propertyAsBytes == nil {
		return shim.Error(fmt.Sprintf("Could not locate property %s.", argProperty))
	}
	property := Property{}
	json.Unmarshal(propertyAsBytes, &property)

	if property.Holder != argCurrentHolder {
		return shim.Error(fmt.Sprintf("Holder %s is invalid for property %s.", argCurrentHolder, argProperty))
	}

	priceValue, err := strconv.Atoi(argPrice)
	
	errorMsg := doPayment(APIstub, argNewHolder, argCurrentHolder, priceValue)
	
	if errorMsg != "" {
		return shim.Error(errorMsg)
	}

	property.Holder = argNewHolder
	propertyAsBytes, _ = json.Marshal(property)
	err = APIstub.PutState(args[0], propertyAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to change property holder: %s", args[0]))
	}
	return shim.Success(nil)
}


/**
 * The pay Method
 * pay(from, to, value)
 * from is the source wallet
 * to is the target wallet
 * 
 */
func (s *SmartContract) pay(APIstub shim.ChaincodeStubInterface) sc.Response {
	_, args := APIstub.GetFunctionAndParameters()

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments for pay. Expecting 3 - (from, to, value)")
	}
	value, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid value for payment.")
	}
	errorMsg := doPayment(APIstub, args[0], args[1], value)
	if errorMsg != "" {
		return shim.Error(errorMsg)
	}
	return shim.Success(nil);
}


/**
 * The doPayment method
 * INTERNAL to this smart contract only.
 * Usage: doPayment(from, to, value)
 * from is the source wallet
 * to is the target wallet
 * value is the exchange value
 * Returns an empty string if the operation is OK.
 * Returns a error message case if:
 * - invalid source
 * - invalid target
 * - insufficient funds from source
 * This method updates BOTH from and to wallets
 */
func doPayment(APIstub shim.ChaincodeStubInterface, from string, to string, value int) (string) {
	// from = to?
	if from == to {
		return fmt.Sprintf("Source wallet and target wallet are equal, %s?", from)
	}
		
	// get SOURCE wallet
	walletFromAsBytes, _ := APIstub.GetState(from)
	if walletFromAsBytes == nil {
		return fmt.Sprintf("Could not locate source wallet for %s", from)
	}
	walletFrom := Wallet{}
	json.Unmarshal(walletFromAsBytes, &walletFrom)
	
	// SOURCE is Active?
	if walletFrom.Status != "Active" {
		return fmt.Sprintf("Wallet for %s is not active anymore.", from)
	}
	
	// get TARGET wallet
	walletToAsBytes, _ := APIstub.GetState(to)
	if walletToAsBytes == nil {
		return fmt.Sprintf("Could not locate target wallet for %s", to)
	}
	walletTo := Wallet{}
	json.Unmarshal(walletToAsBytes, &walletTo)
	
	// TARGET is Active?
	if walletTo.Status != "Active" {
		return fmt.Sprintf("Wallet for %s is not active anymore.", to)
	}
	
	// transfer resources
	if walletTo.Value < value {
		return fmt.Sprintf("Insufficient funds to transfer %d from %s to %s", value, from, to)
	}

	// save both wallets
	x := Wallet{Holder: walletFrom.Holder, Value:  walletFrom.Value - value, Status: "Active"}
	xx, _ := json.Marshal(x)
	err := APIstub.PutState(walletFrom.Holder, xx)
	if err != nil {
		return fmt.Sprintf("Failed to transfer value FROM %s to %s", from, to)
	}
	

	y, _ := json.Marshal(Wallet{Holder: walletTo.Holder, Value:  walletTo.Value + value, Status: "Active"})
	err = APIstub.PutState(walletTo.Holder, y)
	if err != nil {
		return fmt.Sprintf("Failed to transfer value from %s TO %s", from, to)
	}
	return ""
}

/**
 * The bankrupt method.
 * bankrupt(who)
 * Any remaining funds will go to the Bank.
 */
func (s *SmartContract) bankrupt(APIstub shim.ChaincodeStubInterface) sc.Response {
	_, args := APIstub.GetFunctionAndParameters()

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments for pay. Expecting 1 - (who)")
	}
	who := args[0]
	
	if who == "Bank" {
		return shim.Error("Bankrupt BANK is not allowed.")
	}

	walletAsBytes, _ := APIstub.GetState(who)
	if walletAsBytes == nil {
		return shim.Error(fmt.Sprintf("Could not locate wallet for %s", who))
	}
	var wallet Wallet
	json.Unmarshal(walletAsBytes, &wallet)

	if wallet.Value > 0 {
		errorMsg := doPayment(APIstub, who, "Bank", wallet.Value)
		if errorMsg != "" {
			return shim.Error(errorMsg)
		}
	}
	
	wallet.Status = "Inactive"
	w, _ := json.Marshal(wallet)
	err := APIstub.PutState(who, w)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to bankrupt %s", who))
	}

	return shim.Success(nil);
}

/*
 * main function
 * calls the Start function 
 * The main function starts the chaincode in the container during instantiation.
 */
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	} else {
                fmt.Printf("Success creating new Smart Contract")
        }
}