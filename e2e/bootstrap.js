/* Import the puppeteer and expect functionality of chai library for configuring the Puppeteer */
const puppeteer = require('puppeteer');
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
    global.browser = await puppeteer.launch(opts);
});

/* call the function after puppeteer done testing */
after ( () => {
    global.browser.close();
});
