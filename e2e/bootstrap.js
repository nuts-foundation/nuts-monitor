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
    const dockerProcess = process.spawn('docker',["run", "-p", "1323:1323", "-e", "NUTS_STRICTMODE=false", "-e", "NUTS_NETWORK_ENABLETLS=false", "-e", "NUTS_AUTH_CONTRACTVALIDATORS=dummy", "nutsfoundation/nuts-node"])

    // Wait for the container to output "STARTED" in its stdout
    await Promise.race([
        new Promise((resolve, reject) => {
            dockerProcess.stdout.on('data', data => {
                const output = data.toString();
                // uncomment for debugging
                // console.log(output); // Optional: log the container output for debugging
                if (output.includes('Started HTTP')) {
                    resolve();
                }
            });
            dockerProcess.stderr.on('data', data => {
                const output = data.toString();
                // uncomment for debugging
                // console.log(output); // Optional: log the container output for debugging
                if (output.includes('Started HTTP')) {
                    resolve();
                }
            });
        }),
        new Promise((resolve, reject) => {
            setTimeout(() => {
                reject(new Error('Timed out waiting for Docker container to start'));
            }, 5000);
        })
    ]);

    global.dockerProcess = dockerProcess;
});

after( async () => {
    global.browser.close();
    const dockerProcess = global.dockerProcess;
    if (dockerProcess) {
        dockerProcess.kill('SIGINT');
        await new Promise(resolve => {
            dockerProcess.on('exit', resolve);
        });
        global.dockerProcess = null;
    }
})
