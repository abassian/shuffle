package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/abassian/shuffle/src/service/templates"
	"github.com/abassian/shuffle/src/state"
	"github.com/sirupsen/logrus"
)

/*
GET /account/{address}
example: /account/0x50bd8a037442af4cdf631495bcaa5443de19685d
returns: JSON JsonAccount

This endpoint should be used to fetch information about ANY account as opposed
to the /accounts/ endpoint which only returns information about accounts for which
the private key is known and managed by the shuffle Service.
*/
func accountHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	param := r.URL.Path[len("/account/"):]
	m.logger.WithField("param", param).Debug("GET account")
	address := common.HexToAddress(param)
	m.logger.WithField("address", address.Hex()).Debug("GET account")

	balance := m.state.GetBalance(address)
	nonce := m.state.GetNonce(address)
	code := hexutil.Encode(m.state.GetCode(address))
	if code == "0x" {
		code = ""
	} else {
		m.logger.WithField("code", code).Debug("GET account")
	}

	account := JsonAccount{
		Address: address.Hex(),
		Balance: balance,
		Nonce:   nonce,
		Code:    code,
	}

	js, err := json.Marshal(account)
	if err != nil {
		m.logger.WithError(err).Error("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

/*
GET /accounts
returns: JSON JsonAccountList

This endpoint returns the list of accounts CONTROLLED by the shuffle Service.
These are accounts for which the Service has the private keys and on whose behalf
it can sign transactions. The list of accounts controlled by the evm-service is
contained in the Keystore directory defined upon launching the shuffle application.
*/
func accountsHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	m.logger.Debug("GET accounts")

	var al JsonAccountList

	for _, account := range m.keyStore.Accounts() {
		balance := m.state.GetBalance(account.Address)
		nonce := m.state.GetNonce(account.Address)
		code := hexutil.Encode(m.state.GetCode(account.Address))
		if code == "0x" {
			code = ""
		}
		al.Accounts = append(al.Accounts,
			JsonAccount{
				Address: account.Address.Hex(),
				Balance: balance,
				Nonce:   nonce,
				Code:    code,
			})
	}

	js, err := json.Marshal(al)
	if err != nil {
		m.logger.WithError(err).Error("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

/*
POST /call
data: JSON SendTxArgs
returns: JSON JsonCallRes

This endpoints allows calling SmartContract code for READONLY operations. These
calls will NOT modify the EVM state.

The data does NOT need to be signed.
*/
func callHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	m.logger.WithField("request", r).Debug("POST call")

	decoder := json.NewDecoder(r.Body)
	var txArgs SendTxArgs
	err := decoder.Decode(&txArgs)
	if err != nil {
		m.logger.WithError(err).Error("Decoding JSON txArgs")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	callMessage, err := prepareCallMessage(txArgs, m.keyStore)
	if err != nil {
		m.logger.WithError(err).Error("Converting to CallMessage")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := m.state.Call(*callMessage)
	if err != nil {
		m.logger.WithError(err).Error("Executing Call")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := JsonCallRes{Data: hexutil.Encode(data)}
	js, err := json.Marshal(res)
	if err != nil {
		m.logger.WithError(err).Error("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

/*
POST /tx
data: JSON SendTxArgs
returns: JSON JsonTxRes

This endpoints allows calling SmartContract code for NON-READONLY operations.
These operations can MODIFY the EVM state.

The data does NOT need to be SIGNED. In fact, this endpoint is meant to be used
for transactions whose originator is an account CONTROLLED by the shuffle
Service (ie. present in the Keystore).

The Nonce field is not necessary either since the Service will fetch it from the
State.

This is an ASYNCHRONOUS operation. It will return the hash of the transaction that
was SUBMITTED to shuffle but there is no guarantee that the transactions will
get applied to the State.

One should use the /receipt endpoint to retrieve the corresponding receipt and
verify if/how the State was modified.
*/
func transactionHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	m.logger.WithField("request", r).Debug("POST tx")

	decoder := json.NewDecoder(r.Body)
	var txArgs SendTxArgs
	err := decoder.Decode(&txArgs)
	if err != nil {
		m.logger.WithError(err).Error("Decoding JSON txArgs")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	tx, err := prepareTransaction(txArgs, m.state, m.keyStore)
	if err != nil {
		m.logger.WithError(err).Error("Preparing Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.logger.WithFields(logrus.Fields{
		"hash":     tx.Hash().Hex(),
		"to":       tx.To(),
		"payload":  fmt.Sprintf("%x", tx.Data()),
		"gas":      tx.Gas(),
		"gasPrice": tx.GasPrice(),
		"nonce":    tx.Nonce(),
		"value":    tx.Value(),
	}).Debug("Service decoded tx")

	if err := m.state.CheckTx(tx); err != nil {
		m.logger.WithError(err).Error("Checking Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		m.logger.WithError(err).Error("Encoding Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.logger.Debug("submitting tx")
	m.submitCh <- data
	m.logger.Debug("submitted tx")

	res := JsonTxRes{TxHash: tx.Hash().Hex()}
	js, err := json.Marshal(res)
	if err != nil {
		m.logger.WithError(err).Error("Marshalling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

/*
POST /rawtx
data: STRING Hex representation of the raw transaction bytes
	  ex: 0xf8620180830f4240946266b0dd0116416b1dacf36...
returns: JSON JsonTxRes

This endpoint allows sending NON-READONLY transactions ALREADY SIGNED. The client
is left to compose a transaction, sign it and RLP encode it. The resulting bytes,
represented as a Hex string is passed to this method to be forwarded to the EVM.

This allows executing transactions on behalf of accounts that are NOT CONTROLLED
by the shuffle service.

Like the /tx endpoint, this is an ASYNCHRONOUS operation and the effect on the
State should be verified by fetching the transaction' receipt.
*/
func rawTransactionHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	m.logger.WithField("request", r).Debug("POST rawtx")

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		m.logger.WithError(err).Error("Reading request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m.logger.WithField("body", body)

	sBody := string(body)
	m.logger.WithField("body (string)", sBody).Debug()
	rawTxBytes, err := hexutil.Decode(sBody)
	if err != nil {
		m.logger.WithError(err).Error("Reading raw tx from request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m.logger.WithField("raw tx bytes", rawTxBytes).Debug()

	var t ethTypes.Transaction
	if err := rlp.Decode(bytes.NewReader(rawTxBytes), &t); err != nil {
		m.logger.WithError(err).Error("Decoding Transaction")
		return
	}

	m.logger.WithFields(logrus.Fields{
		"hash":     t.Hash().Hex(),
		"to":       t.To(),
		"payload":  fmt.Sprintf("%x", t.Data()),
		"gas":      t.Gas(),
		"gasPrice": t.GasPrice(),
		"nonce":    t.Nonce(),
		"value":    t.Value(),
	}).Debug("Service decoded tx")

	if err := m.state.CheckTx(&t); err != nil {
		m.logger.WithError(err).Error("Checking Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.logger.Debug("submitting tx")
	m.submitCh <- rawTxBytes
	m.logger.Debug("submitted tx")

	res := JsonTxRes{TxHash: t.Hash().Hex()}
	js, err := json.Marshal(res)
	if err != nil {
		m.logger.WithError(err).Error("Marshalling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

/*
GET /tx/{tx_hash}
ex: /tx/0xbfe1aa80eb704d6342c553ac9f423024f448f7c74b3e38559429d4b7c98ffb99
returns: JSON JsonReceipt

This endpoint allows to retrieve the EVM receipt of a specific transactions if it
exists. When a transaction is applied to the EVM , a receipt is saved to allow
checking if/how the transaction affected the state. This is where one can see such
information as the address of a newly created contract, how much gas was use and
the EVM Logs produced by the execution of the transaction.
*/
func transactionReceiptHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	param := r.URL.Path[len("/tx/"):]
	txHash := common.HexToHash(param)
	m.logger.WithField("tx_hash", txHash.Hex()).Debug("GET tx")

	tx, err := m.state.GetTransaction(txHash)
	if err != nil {
		m.logger.WithError(err).Error("Getting Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	receipt, err := m.state.GetReceipt(txHash)
	if err != nil {
		m.logger.WithError(err).Error("Getting Receipt")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	signer := ethTypes.NewEIP155Signer(big.NewInt(1))
	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		m.logger.WithError(err).Error("Getting Tx Sender")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonReceipt := JsonReceipt{
		Root:              common.BytesToHash(receipt.PostState),
		TransactionHash:   txHash,
		From:              from,
		To:                tx.To(),
		GasUsed:           receipt.GasUsed,
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		ContractAddress:   receipt.ContractAddress,
		Logs:              receipt.Logs,
		LogsBloom:         receipt.Bloom,
		Status:            receipt.Status,
	}

	if receipt.Logs == nil {
		jsonReceipt.Logs = []*ethTypes.Log{}
	}

	js, err := json.Marshal(jsonReceipt)
	if err != nil {
		m.logger.WithError(err).Error("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

/*
GET /info
returns: JSON (depends on underlying consensus system)

Info returns information about the consensus system. Each consensus system that
plugs into shuffle must implement an Info function.
*/
func infoHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	m.logger.Debug("GET info")

	stats, err := m.getInfo()
	if err != nil {
		m.logger.WithError(err).Error("Getting Info")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(stats)
	if err != nil {
		m.logger.WithError(err).Error("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

/*
GET /html/info
returns: HTML version of info

Info returns information about the consensus system. Each consensus system that
plugs into shuffle must implement an Info function.
*/
func htmlInfoHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	m.logger.Debug("GET html/info")

	stats, err := m.getInfo()
	if err != nil {
		m.logger.WithError(err).Error("Getting Info")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t := template.New("index")        //name of the template is index
	t, err = t.Parse(templates.Index) // parsing of template string
	if err != nil {
		m.logger.WithError(err).Error("Parsing template")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, stats)
}

/*
GET /contract
returns: JSON
Returns details of the poa smart contract .   *** DEPRECATED ***
*/
func contractHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	m.logger.Debug("GET contract")

	var al JsonContractList

	al.Contracts = append(al.Contracts, JsonContract{
		Address: common.HexToAddress("abbaabbaabbaabbaabbaabbaabbaabbaabbaabba"),
		ABI:     `[{"constant":true,"inputs":[{"name":"_publicKey","type":"bytes32"}],"name":"checkAuthorisedPublicKey","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_address","type":"address"}],"name":"dev_isGenesisWhitelisted","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"dev_27","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[{"name":"_address","type":"address"}],"name":"checkAuthorised","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_nomineeAddress","type":"address"}],"name":"getNoVoteCount","outputs":[{"name":"count","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"whiteList","outputs":[{"name":"person","type":"address"},{"name":"flags","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"dev_getSender","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_nomineeAddress","type":"address"},{"name":"_idx","type":"uint256"}],"name":"getYesVoterFromIdx","outputs":[{"name":"voter","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_address","type":"address"}],"name":"getNomineeElection","outputs":[{"name":"nominee","type":"address"},{"name":"proposer","type":"address"},{"name":"yesVotes","type":"uint256"},{"name":"noVotes","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_nomineeAddress","type":"address"},{"name":"_idx","type":"uint256"}],"name":"getNoVoterFromIdx","outputs":[{"name":"voter","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"idx","type":"uint256"}],"name":"getNomineeElectionFromIdx","outputs":[{"name":"nominee","type":"address"},{"name":"proposer","type":"address"},{"name":"yesVotes","type":"uint256"},{"name":"noVotes","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getWhiteListCount","outputs":[{"name":"count","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getNomineeCount","outputs":[{"name":"count","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"dev_getWhitelistCount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"idx","type":"uint256"}],"name":"getWhiteListAddressFromIdx","outputs":[{"name":"WhiteListAddress","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"idx","type":"uint256"}],"name":"getNomineeAddressFromIdx","outputs":[{"name":"NomineeAddress","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_nomineeAddress","type":"address"}],"name":"getYesVoteCount","outputs":[{"name":"count","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_address","type":"address"}],"name":"getMoniker","outputs":[{"name":"moniker","type":"bytes32"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_nomineeAddress","type":"address"},{"name":"_accepted","type":"bool"}],"name":"castNomineeVote","outputs":[{"name":"decided","type":"bool"},{"name":"voteresult","type":"bool"}],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[],"name":"dev_getGenesisWhitelist0","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":false,"inputs":[],"name":"init","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"_address","type":"address"}],"name":"dev_isWhitelisted","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_address","type":"address"}],"name":"dev_getCurrentNomineeVotes","outputs":[{"name":"yes","type":"uint256"},{"name":"no","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_nomineeAddress","type":"address"},{"name":"_moniker","type":"bytes32"}],"name":"submitNominee","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"inputs":[{"name":"_moniker","type":"bytes32"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"name":"_nominee","type":"address"},{"indexed":false,"name":"_yesVotes","type":"uint256"},{"indexed":false,"name":"_noVotes","type":"uint256"},{"indexed":true,"name":"_accepted","type":"bool"}],"name":"NomineeDecision","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"_nominee","type":"address"},{"indexed":true,"name":"_voter","type":"address"},{"indexed":false,"name":"_yesVotes","type":"uint256"},{"indexed":false,"name":"_noVotes","type":"uint256"},{"indexed":true,"name":"_accepted","type":"bool"}],"name":"NomineeVoteCast","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"_nominee","type":"address"},{"indexed":true,"name":"_proposer","type":"address"}],"name":"NomineeProposed","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"_address","type":"address"},{"indexed":true,"name":"_moniker","type":"bytes32"}],"name":"MonikerAnnounce","type":"event"}]`,
	})

	js, err := json.Marshal(al)
	if err != nil {
		m.logger.WithError(err).Error("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

/*
GET /poa
returns: JSON
Returns details of the poa smart contract . Replaces /contract
*/
func poaHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	m.logger.Debug("GET poa")

	var al JsonContract

	al = JsonContract{
		Address: common.HexToAddress(m.state.GetAuthorisingAccount()),
		ABI:     m.state.GetAuthorisingAbi(),
	}

	js, err := json.Marshal(al)
	if err != nil {
		m.logger.WithError(err).Error("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

/*
GET /genesis
returns: JSON Genesis

This endpoint returns the content of the genesis.json file.
*/
func genesisHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	m.logger.Debug("GET genesis")

	genesis, err := m.state.GetGenesis()
	if err != nil {
		m.logger.WithError(err).Error("Getting Genesis")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(genesis)
	if err != nil {
		m.logger.WithError(err).Error("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

//------------------------------------------------------------------------------
func prepareCallMessage(args SendTxArgs, ks *keystore.KeyStore) (*ethTypes.Message, error) {
	var err error
	args, err = prepareSendTxArgs(args)
	if err != nil {
		return nil, err
	}

	//Todo set default from

	// Create Call Message
	// Set gasPrice and value to 0 because this is a readonly operation
	msg := ethTypes.NewMessage(args.From,
		args.To,
		0,
		big.NewInt(0),
		args.Gas,
		big.NewInt(0),
		common.FromHex(args.Data),
		false)

	return &msg, nil

}

func prepareTransaction(args SendTxArgs, state *state.State, ks *keystore.KeyStore) (*ethTypes.Transaction, error) {
	var err error
	args, err = prepareSendTxArgs(args)
	if err != nil {
		return nil, err
	}

	if args.Nonce == nil {
		args.Nonce = new(uint64)
		*args.Nonce = state.GetPoolNonce(args.From)
	}

	var tx *ethTypes.Transaction
	if args.To == nil {
		tx = ethTypes.NewContractCreation(*args.Nonce,
			args.Value,
			args.Gas,
			args.GasPrice,
			common.FromHex(args.Data))
	} else {
		tx = ethTypes.NewTransaction(*args.Nonce,
			*args.To,
			args.Value,
			args.Gas,
			args.GasPrice,
			common.FromHex(args.Data))
	}

	signer := ethTypes.NewEIP155Signer(big.NewInt(1))

	account, err := ks.Find(accounts.Account{Address: args.From})
	if err != nil {
		return nil, err
	}
	signature, err := ks.SignHash(account, signer.Hash(tx).Bytes())
	if err != nil {
		return nil, err
	}
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func prepareSendTxArgs(args SendTxArgs) (SendTxArgs, error) {
	if args.Gas == 0 {

	}
	if args.GasPrice == nil {
		args.GasPrice = big.NewInt(0)
	}
	if args.Value == nil {
		args.Value = big.NewInt(0)
	}
	return args, nil
}
