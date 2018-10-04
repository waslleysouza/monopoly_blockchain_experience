var express = require('express');
var router = express.Router();
var request = require('request');

var channel = process.env.CHANNEL;
var chaincode = process.env.CHAINCODE;
var chaincodeVer = process.env.CHAINCODE_VERSION;
var method = null;
var player = null;
var property = null;

router.get('/', function (req, res, next) {
    res.render('query', { 
        channel: channel,
        chaincode: chaincode,
        chaincodeVer: chaincodeVer,
        method: method,
        player: player,
        property: property,
        transactions: [],
        message: null });
});

router.post('/', function (req, res, next) {
    
    console.log(req.body);

    var args = [];
    if (req.body.method == "queryWallet" || req.body.method == "queryWalletHistory") {
        args = [req.body.player];

    } else if (req.body.method == "queryProperty" || req.body.method == "queryPropertyHistory") {
        args = [req.body.property];
    }

    var json = {
        "channel": req.body.channel,
        "chaincode": req.body.chaincode,
        "chaincodeVer": req.body.chaincodeVer,
        "method": req.body.method,
        "args": args
    };
    
    // Configure the request
    var options = {
        url: process.env.URL_QUERY,
        method: "POST",
        json: json,
        proxy: ""
    };

    // Start the request
    request(options, function (error, response, body) {
        if (error) {
            console.error("Error: " + error);
        }

        console.log(body);

        var transactions = [];
        if (body.returnCode == "Success") {
            var result = JSON.parse(body.result);

            if (result instanceof Array) {
                transactions = result;
            } else {
                transactions.push(result);
            }

        } else {
            var message = JSON.stringify(json, undefined, 2) + "\n\Error!\n\n" + body.info;
        }
        
        res.render('query', { 
            channel: req.body.channel,
            chaincode: req.body.chaincode,
            chaincodeVer: req.body.chaincodeVer,
            method: req.body.method,
            player: req.body.player,
            property: req.body.property,
            transactions: transactions,
            message: message });
    });
});

module.exports = router;