document.body.addEventListener('htmx:afterRequest', function (event) {
    if (event.detail.xhr.status == 200) {
        window.location.reload();
    }
});