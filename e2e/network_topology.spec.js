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
        await page.waitForSelector('svg g g'); // Wait for a node to appear
        const nodeCount = await page.$$eval('svg g g', nodes => nodes.length);
        expect(nodeCount).to.equal(2); // Check that only one circle has been added
    });
});