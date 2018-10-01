var fields = $('.args');
var clazz = $('#method').val();
fields.each(function() {
    var field = $(this);
    if (field.hasClass(clazz)) field.show();
    else field.hide();
});

$('#method').on('change', function() {
    var fields = $('.args').hide();
    var clazz = this.value;
	fields.each(function() {
        var field = $(this);
        if (field.hasClass(clazz)) { 
            field.show();
            // Select option index 0
            field.find(".form-control option:eq(0)").prop('selected', true); 
        }
        else field.hide();
    });
});