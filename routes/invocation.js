var express = require('express');
var router = express.Router();
var request = require('request');

var channel = process.env.CHANNEL;
var chaincode = process.env.CHAINCODE;
var chaincodeVer = process.env.CHAINCODE_VERSION;
var method = null;
var player = null;
var player2 = null;
var property = null;
var value = null;

router.get('/', function (req, res, next) {
    res.render('invocation', { 
        channel: channel,
        chaincode: chaincode,
        chaincodeVer: chaincodeVer,
        method: method,
        player: player,
        player2: player2,
        property: property,
        value: value,
        message: null });
});

router.post('/', function (req, res, next) {
    
    console.log(req.body);

    var args = [];
    if (req.body.method == "transferProperty") {
        args = [req.body.property, req.body.player, req.body.player2, req.body.value];

    } else if (req.body.method == "pay") {
        args = [req.body.player, req.body.player2, req.body.value];
    
    } else if (req.body.method == "bankrupt") {
        args = [req.body.player];
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
        url: process.env.URL_INVOCATION,
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

        var message = "";
        if (body.returnCode == "Success") {
            message = JSON.stringify(json, undefined, 2) + "\n\nTransaction added!\n\ntransactionID: " + body.transactionID;

        } else {
            message = JSON.stringify(json, undefined, 2) + "\n\nError!\n\n" + body.info;
        }
        
        res.render('invocation', { 
            channel: req.body.channel,
            chaincode: req.body.chaincode,
            chaincodeVer: req.body.chaincodeVer,
            method: req.body.method,
            player: req.body.player,
            player2: req.body.player2,
            property: req.body.property,
            value: req.body.value,
            message: message });
    });
});

module.exports = router;