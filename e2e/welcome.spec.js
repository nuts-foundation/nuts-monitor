describe('Landing page loading', async  () => {
    let page;

    before(async () => { /* before hook for mocha testing */
        page = await browser.newPage();
        const [response] = await Promise.all([
            page.goto("http://localhost:1323", {timeout:0}),
            page.waitForNavigation({timeout:0}),
        ]);
    });

    after(async function () { /* after hook for mocha testing */
        await page.close();
    });

    it('should display title', async () => {
        expect(await page.title()).to.eql('Nuts monitor');
    });
});