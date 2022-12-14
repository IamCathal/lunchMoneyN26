const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun",
    "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"
];
let settingsPanelIsVisible = false;

window.onload = function() {
    giveSwayaaangBordersToItems()
    initVariablesIfNotSet()

    checkStatus().then((res) => {
        setInterval(checkStatus, 10000)
    })
    fillInDaysInput()
    fillInRecentTransactions()
}

function initVariablesIfNotSet() {
    getStorageVariable("transactions").then((transactions) => {
        if (!Array.isArray(transactions)) {
            let transactionsArr = []
            browser.storage.local.set({
                "transactions": transactionsArr
            }).then(() => {
                console.log("Successfully set empty transactions")
            }, (err) => {
                console.error(`Failed to set empty transactions: ${err}`)
            })
        }
    }, (err) => {
        console.error(`failed to get transactions: ${err}`)
    })
    getStorageVariable("backendURL").then((backendURL) => {
        if (backendURL == "" || backendURL === undefined) {
            browser.storage.local.set({
                "backendURL": "empty"
            }).then(() => {
                console.log("Successfully set empty backendURL")
            }, (err) => {
                console.error(`Failed to set empty backendURL: ${err}`)
            })
        } else {
            console.log(`backendURL was not empty: ${backendURL}`)
        }
    }, (err) => {
        console.error(`failed to get backendURL: ${err}`)
    })
}

function giveSwayaaangBordersToItems() {
    document.getElementById("settingsButton").style = swayaaangBorders(0.6)
    // document.getElementById("importTransactions").style += swayaaangBorders(0.6)   
}

function swayaaangBorders(borderRadius) {
    const borderArr = [
        `border-top-right-radius: ${borderRadius}rem;`, 
        `border-bottom-right-radius: ${borderRadius}rem;`,
        `border-top-left-radius: ${borderRadius}rem;`,
        `border-bottom-left-radius: ${borderRadius}rem;`,
    ]

    let borderRadiuses = "";
    for (let k = 0; k < 4; k++) {
        const randNum = Math.floor(Math.random() * 2)
        if (randNum % 2 == 0) {
            borderRadiuses += borderArr[k]
        }
    } 
    return borderRadiuses
}

function getMostRecentTransactions(numTransactions) {
    return new Promise((resolve, reject) => {
        getStorageVariable("transactions", []).then((res) => {
            if (res === undefined) {
                resolve([])
            } else if (res.length >= numTransactions) {
                resolve(res.slice(numTransactions))
            } else {
                resolve(res)
            }
        }, (err) => {
            reject(err)
        })
    })
}

function fillInDaysInput() {
    getStorageVariable("daysInput", []).then((res) => {
        if (res === undefined) {
            document.getElementById("daysInput").value = "1"
        } else {
            document.getElementById("daysInput").value = res
        }
    }, (err) => {
        console.error(`err getting daysInput:${err}`)
    })
}

function fillInRecentTransactions() {
    let outputHTML = "";

    getMostRecentTransactions(5).then((transactions) => {
        for (let i = 0; i < transactions.length; i++) {
            const curr = transactions[i];
            const transactionDate = new Date(curr.currTime);
            let dateAtBeginningOfScan = new Date(curr.currTime);
            dateAtBeginningOfScan = new Date(dateAtBeginningOfScan.setDate(dateAtBeginningOfScan.getDate() - curr.daysLookedUp))
    
            // Add a bottom margin if its not the last element
            const bottomMargin = i == (transactions.length - 1) ? '' : 'margin-bottom: 0.3rem;';
    
            outputHTML += `
            <td>
            </td>
            <div style="${bottomMargin} padding: 0.2rem 0.4rem 0.2rem 0.4rem; border: 1px solid grey; ${swayaaangBorders(0.55)}">
                <table>
                    <tr style="text-align: center; width: 100%; float: center">
    
                        <td style="width: 1.5rem">
                            <img
                                src="https://i.imgur.com/vNnbGKp.png"
                                width="100%"
                                style="filter: invert()"
                            >   
                        </td>
    
                        <td>
                        </td>
    
                        <td style="font-weight: bold; width: 20%; margin-left: 0.4vh; text-align: left">
                            ${curr.n26FoundTransactions}
                        </td>
    
                        <td style="width: 20%">
                        </td>
    
                        <td style="font-weight: bold; width: 20%; text-align: right ">
                            ${curr.lunchMoneyInseredTransactions}
                        </td>
    
                        <td>
                        </td>
    
                        <td style="width: 50%;float: right">
                            <img
                                src="https://i.imgur.com/aD0grat.png"
                                width="60%"
                            >
                        </td>
    
                    </tr>
                </table>
                <div style="text-align: center">
                    <div style="display: inline; font-size: 0.65rem">
                        ${getDaysApartOutputString(dateAtBeginningOfScan, transactionDate)}
                    </div>
                </div>
            </div>
            `
        }
        document.getElementById("transactionBay").innerHTML = outputHTML;
    }, (err) => {
        console.error(err)
    })
}

document.getElementById("importTransactionsButton").addEventListener("click", function(e){
    e.preventDefault();
    getInfoForRequest().then((info) => {
        const backendURLObject = new URL(info.backendURL)
        const ws = new WebSocket(`ws://${backendURLObject.host}/ws/transactions?days=${info.daysInput}&apikey=${info.apiKey}`);

        ws.onopen = function(e) {
            createRequestStatusBox() 
        };

        ws.onmessage = function(event) {
            handleNewWsMessage(event)
        }

        ws.onclose = function(event) {
            if (event.wasClean) {
                console.log(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
            } else {
                console.log('[close] Connection died');
            }
            hideCreateStatusRequestBox() 
        };
        }, (err) => {
            console.error(`failed to get backendURL: ${err}`)
    })
});

function handleNewWsMessage(event) {
    const res = JSON.parse(event.data)
    console.log(res)

    // This is not best practice teehee
    if (res.msg == "N26 has been authorized") {
        document.getElementById("waitingFor2FAIcon").classList.remove("skeleton");
    }
    else if (res.msg == "Retrieved transactions from N26") {
        document.getElementById("retrievingFromN26Icon").classList.remove("skeleton");
    }
    else if (res.msg == "transactions inserted into LunchMoney") {
        document.getElementById("insertingIntoLunchMoneyIcon").classList.remove("skeleton");
    }
    else if (res.msg == "transaction finished") {
        document.getElementById("savingSummaryIcon").classList.remove("skeleton");
        const newTransaction = res.summarystats;

        getStorageVariable("transactions").then((transactions) => {
            if (!Array.isArray(transactions)) {
                transactions = []
            }
            transactions.unshift(newTransaction)
            console.log(`New transaction list:`)
            console.log(transactions)

            browser.storage.local.set({
                "transactions": transactions
            }).then(() => {
                console.log("Successfully set new transactions")
                fillInRecentTransactions()
            }, (err) => {
                console.error(`Failed to set new transactions: ${err}`)
            })
        }, (err) => {
            console.error(err)
        })
    }
}

function getDaysApartOutputString(first, second) {
    if (first.getMonth() == second.getMonth()) {
        return `${first.getDate()} - ${second.getDate()} ${monthNames[first.getMonth()]}`
    } else {
        return `${first.getDate()} ${monthNames[first.getMonth()]} - ${second.getDate()} ${monthNames[second.getMonth()]}`
    }
}

function dateString(date) {
    return date.getDate() + " " + monthNames[date.getMonth()]
}

function checkStatus() {
    return new Promise((resolve, reject) => {
        getInfoForRequest().then((info) => {
            if (info.backendURL === "empty") {
                resolve()
            }
            console.log(`getting status from ${info.backendURL}/status`)
            fetch(`${info.backendURL}/status`, {
                method: 'POST',
                mode: 'cors',
                headers: {
                    'API_KEY': info.apiKey
                }
            })
            .then((res) => res.json())
            .then((data) => JSON.parse(data))
            .then((data) => {
                const currTime = new Date()
                const timeAtServerStartUp = new Date(data.startuptime * 1000);
                document.getElementById("upTime").textContent = timeSince(timeAtServerStartUp)
                document.getElementById("onlineStatusText").textContent = "Online"
                resolve()
            }, (err) => {
                document.getElementById("upTime").textContent = "";
                document.getElementById("onlineStatusText").textContent = "Offline"
                document.getElementById("onlineStatusIndicator").style = "filter: invert(30)"
                reject()
            });
        }, (err) => {
            console.error("failed to get backendURL and apikey")
            reject(err)
        })
    })
}


function createRequestStatusBox() {
    console.log("opening")
    document.getElementById("requestStatusBox").innerHTML = `
    <div style="text-align: center; margin-bottom: 0.0; margin-top: 0.3rem; border: 1px solid grey">
            <table style="font-size: 0.75rem">
                <tr>
                    <td>
                        <span style="margin-left: 0.5vh; margin-right: 0.5vh">
                            <img
                                src="https://i.imgur.com/3Ktp7tL.png"
                                width="4%"
                                style="margin-top: 0.3vh;"
                                id="waitingFor2FAIcon"
                                class="skeleton"
                            >
                        </span>
                        Authorize login within <span id="2FACountdownText">5m</span>
                    </td>
                </tr>
                <tr>
                    <td>
                        <span style="margin-left: 0.5vh; margin-right: 0.5vh">
                            <img
                                src="https://i.imgur.com/3Ktp7tL.png"
                                width="4%"
                                style="margin-top: 0.3vh"
                                id="retrievingFromN26Icon"
                                class="skeleton"
                            >
                        </span>
                        Retrieving from N26
                    </td>
                </tr>
                <tr>
                    <td>
                        <span style="margin-left: 0.5vh; margin-right: 0.5vh">
                            <img
                                src="https://i.imgur.com/3Ktp7tL.png"
                                width="4%"
                                style="margin-top: 0.3vh"
                                id="insertingIntoLunchMoneyIcon"
                                class="skeleton"
                            >
                        </span>
                        Inserting into LunchMoney
                    </td>
                </tr>
                <tr>
                    <td>
                        <span style="margin-left: 0.5vh; margin-right: 0.5vh">
                            <img
                                src="https://i.imgur.com/3Ktp7tL.png"
                                width="4%"
                                style="margin-top: 0.3vh"
                                id="savingSummaryIcon"
                                class="skeleton"
                            >
                        </span>
                        Saving summary
                    </td>
                </tr>
            </table>
        </div>
    `;
}

function hideCreateStatusRequestBox() {
    console.log("closing")
    document.getElementById("requestStatusBox").innerHTML = ``;
}

// https://stackoverflow.com/a/3177838
// no point in rewriting this
function timeSince(date) {
    let seconds = Math.floor((new Date() - date) / 1000);
    let interval = seconds / 31536000;

    if (interval > 1) {
        return Math.floor(interval) + "y";
    }
    interval = seconds / 2592000;
    if (interval > 1) {
        return Math.floor(interval) + "m";
    }
    interval = seconds / 86400;
    if (interval > 1) {
        return Math.floor(interval) + "d";
    }
    interval = seconds / 3600;
    if (interval > 1) {
        return Math.floor(interval) + "h";
    }
    interval = seconds / 60;
    if (interval > 1) {
        return Math.floor(interval) + "m";
    }
    return Math.floor(seconds) + "s";
}

document.getElementById("daysInput").addEventListener("change", (ev) => {
    const inputNum = parseInt(document.getElementById("daysInput").value)
    console.log(inputNum)
    if (Number.isInteger(parseInt(inputNum))) {
        if (inputNum > 0 && inputNum < 365) {
            browser.storage.local.set({
                "daysInput": inputNum
            }).then((res) => {
                console.log("Successfully set daysInput")
            }, (err) => {
                console.error(`err setting daysInput: ${err}`)
            })
        }
    } else {
        console.log("is not in range")
    }
})

document.getElementById("settingsButton").addEventListener("click", (event) => {
    settingsPanelButtonClicked()
});

function settingsPanelButtonClicked() {
    if (settingsPanelIsVisible) {
        document.getElementById("settingsPanel").innerHTML = ``
        settingsPanelIsVisible = false
        return
    }

    settingsPanelIsVisible = true
    getInfoForRequest().then((info) => {
        document.getElementById("settingsPanel").innerHTML = `
        <div style="padding-top: 0; padding-bottom: 0; padding-left: 0.8vh; padding-right: 0.8vh">
            <hr style="padding: 0; margin: 0; margin-top: 0.4vh"/>
            <table style="text-align: center; width: 100%">
                <tr>
                    <td style="font-size: 0.8rem;">
                        Backend URL
                    </td>
                </tr>
                <tr>
                    <td style="width: 100%">
                        <input 
                            type="text" id="backendURLInput"
                            style="font-size: 0.6rem; background-color: #464646; border: 1px solid grey; color: white; width: 100%"
                            value=${info.backendURL === "empty" ? "" : info.backendURL}
                        >
                    </td>
                </tr>

                <tr>
                    <td style="font-size: 0.8rem;">
                        API Key
                    </td>
                </tr>
                <tr>
                    <td style="width: 100%">
                        <input 
                            type="text" id="apiKeyInput"
                            style="font-size: 0.6rem; background-color: #464646; border: 1px solid grey; color: white; width: 100%"
                            value=${info.apiKey === "empty" ? "" : info.apiKey}
                        >
                    </td>
                </tr>

                <tr>
                    <td style="float: left">
                        <span>
                            <button id="testBackendURL" style="font-size: 3.8vh; width: 14vh">
                                Test
                            </button>   
                        </span>
                        <span id="testBackendStatus" style="font-size: 4vh;">
                            
                        </span>
                    </td>
                </tr>
            </table>
        </div>   
        `
        document.getElementById("testBackendURL").addEventListener("click", (event) => {
            const newURLString = document.getElementById("backendURLInput").value;
            const apiKey = document.getElementById("apiKeyInput").value;
            try {
                const testURL = new URL(newURLString)
            } catch (_) {
                document.getElementById("testBackendStatus").textContent = "Nope"
                return
            }
            console.log(`curl ${newURLString}/status API_KEY: ${apiKey}`)
            fetch(`${newURLString}/status`, {
                method: 'POST',
                mode: 'cors',
                headers: {
                    'API_KEY': apiKey
                }
            })
            .then((res) => res.json())
            .then((data) => JSON.parse(data))
            .then((data) => {
                if (data.status == "operational") {
                    document.getElementById("testBackendStatus").textContent = "Working :)"
                    setBackendURL(newURLString).then(() => {}, (err) => {
                        console.error(`failed to set backendURL: ${err}`)
                    })
                    setApiKey(apiKey).then(() => {}, (err) => {
                        console.error(`failed to set apiKey: ${err}`)
                    })
                } else {
                    console.log("invalid response")
                    console.log(data)
                    document.getElementById("testBackendStatus").textContent = "Nope"
                }
            }, (err) => {
                console.log("err response")
                document.getElementById("testBackendStatus").textContent = "Nope"
            });
        });
    })

}

function getInfoForRequest() {
    return new Promise((resolve, reject) => {
        getStorageVariable("backendURL").then((backendURL) => {
            getStorageVariable("apiKey").then((apiKey) => {
                getStorageVariable("daysInput").then((daysInput) => {
                    resolve({"backendURL": backendURL, "apiKey": apiKey, "daysInput": daysInput})
                }, (err) => {
                    reject(`failed to get daysInput: ${err}`)
                })
            }, (err) => {
                reject(`failed to get apiKey: ${err}`)
            })
        }, (err) => {
            reject(`failed to get backendURL: ${err}`)
        })
    })
}

function setBackendURL(URL) {
    return new Promise((resolve, reject) => {
        browser.storage.local.set({
            "backendURL": URL
        }).then((res) => {
            console.log("Successfully set backend URL")
            resolve()
        }, (err) => {
            reject(err)
        })
    })
}

function setApiKey(apiKey) {
    return new Promise((resolve, reject) => {
        browser.storage.local.set({
            "apiKey": apiKey
        }).then((res) => {
            console.log("Successfully set apiKey")
            resolve()
        }, (err) => {
            reject(err)
        })
    })
}

function getStorageVariable(variable) {
    return new Promise((resolve, reject) => {
        browser.storage.local.get(variable).then(
            (res) => {
                console.log(`Retrieved: '${res[variable]}' type: ${typeof res} from get '${variable}'`)
                resolve(res[variable])
            }, (err) => {
                reject(err)
            }
        )
    })
}