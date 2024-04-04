$(document).ready(function () {
    $("#show_hide_password a, #show_hide_repeat_password a").on('click', function (event) {
        event.preventDefault();
        var input = $(this).closest('.input-group').find('input');
        var img = $(this).closest('.input-group').find('img');

        if (input.attr("type") == "text") {
            input.attr('type', 'password');
            img.attr('src', '/static/resources/images/eye-slash.svg');
        } else if (input.attr("type") == "password") {
            input.attr('type', 'text');
            img.attr('src', '/static/resources/images/eye.svg');
        }
    });
});
