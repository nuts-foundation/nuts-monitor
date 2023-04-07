describe('Network topology loading', async  () => {
    let page;

    before(async () => { /* before hook for mocha testing */
        page = await browser.newPage();
        const [response] = await Promise.all([
            page.goto("http://localhost:1313/#/network_topology", {timeout:0}),
            page.waitForNavigation({timeout:0}),
        ]);
    });

    after(async function () { /* after hook for mocha testing */
        await page.close();
    });

    it('should display a single node', async () => {
        await page.waitForSelector('svg circle'); // Wait for the circle to appear
        const circleCount = await page.$$eval('svg circle', circles => circles.length);
        expect(circleCount).to.equal(1); // Check that only one circle has been added
    });
});