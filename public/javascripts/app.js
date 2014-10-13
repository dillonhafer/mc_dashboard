$(document).on('click', '.button', function(e) {
  e.preventDefault();
  $.post(this.getAttribute('href'));
})
$(document).foundation();