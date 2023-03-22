/* Import the puppeteer and expect functionality of chai library for configuring the Puppeteer */
const puppeteer = require('puppeteer');
const { spawn } = require('child_process')
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
    global.binary = spawn('go', ['run', '.', 'live'], {
        detached: true,
        stdio: 'ignore',
    });
    await new Promise(r => setTimeout(r, 5000));
});

/* call the function after puppeteer done testing */
after ( () => {
    global.browser.close();
    global.binary.kill();
});