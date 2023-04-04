/* Import the puppeteer and expect functionality of chai library for configuring the Puppeteer */
const puppeteer = require('puppeteer');
const process = require('child_process');
const expect = require('chai').expect;

/* configurable options or object for puppeteer */
const opts = {
    headless: true,
    slowMo: 1,
    timeout: 0,
    args: ['--window-size=1600,1200']
}

/* call the before for puppeteer for execute this code before start testing */
before (async () => {
    global.expect = expect;
    global.nuts_node = process.spawn('docker',["run", "-p", "1323:1323", "-e", "NUTS_STRICTMODE=false", "-e", "NUTS_NETWORK_ENABLETLS=false", "-e", "NUTS_AUTH_CONTRACTVALIDATORS=dummy", "nutsfoundation/nuts-node"])
    global.browser = await puppeteer.launch(opts);
});

/* call the function after puppeteer done testing */
after ( () => {
    global.browser.close();
    global.nuts_node.kill();
});
