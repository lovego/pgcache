// vim: set ft=html:
package manage

func listHtml(table string) string {
return `
<!doctype html>
<html>
  <head>
    <meta charset="UTF-8">
    <base target="_blank">
    <style>
table { width: 100%; border-collapse: collapse; }
th,td { padding: 5px 10px; border: 1px dashed gray; }
input { width: 95%; margin-top: 0.5em; padding: 3px 7px; border: 1px solid #ccc; border-radius: 3px; }
input:focus { outline-width: 0; }
    </style>
  </head>
  <body>
    <p>To see a spefic value in data, click at blank area after data link to show input box(click again to hide it).
    Input map keys or slice indexes seperated by "," and press ENTER.
    </p>
    ` + table + `
    <script>
      function createInput(event) {
        var td = event.currentTarget;
        if (td != event.target) return;

        if (td.lastElementChild.tagName != 'INPUT') {
          var input = document.createElement('input');
          input.addEventListener("keyup", function(event) {
            if (event.keyCode !== 13) return;
            input.value = input.value.trim();
            if (input.value == "") return;
            window.open(td.firstElementChild.getAttribute("href") + "?keys=" + input.value)
          });

          td.appendChild(document.createElement('br'));
          td.appendChild(input);
          input.focus();
        } else {
          var style = td.lastElementChild.style;
          if (style.display === '') {
            style.display = 'none';
          } else {
            style.display = '';
            td.lastElementChild.focus();
          }
        }
      }

      (function() {
        var collection = document.querySelectorAll('td.data');
        for (var i = 0; i < collection.length; i++) {
          collection[i].addEventListener('click',  createInput);
        }
      })();
    </script>
  </body>
</html>
`
}
