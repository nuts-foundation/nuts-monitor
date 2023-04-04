describe('Main page loading', async  () => {
    let page;

    before(async () => { /* before hook for mocha testing */
        page = await browser.newPage();
        const [response] = await Promise.all([
            page.goto("http://localhost:1313", {timeout:0}),
            page.waitForNavigation({timeout:0}),
        ]);
    });

    after(async function () { /* after hook for mocha testing */
        await page.close();
    });

    it('should display title', async () => {
        expect(await page.title()).to.eql('Nuts monitor');
    });
    it('should display diagnostics', async () => {
        // Get the div element and check its contents
        const documentsCountDiv = await page.$('#documents_count');
        const textContent = await documentsCountDiv.evaluate(node => node.textContent);
        expect(textContent).to.equal('0');
    });
});