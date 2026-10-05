(function () {
  var rows = document.getElementById('rows');
  var status = document.getElementById('status');
  var refresh = document.getElementById('refresh');

  function render(records) {
    rows.innerHTML = '';
    records.forEach(function (record, index) {
      var tr = document.createElement('tr');
      var n = document.createElement('td');
      n.textContent = String(index + 1);
      var v = document.createElement('td');
      v.textContent = JSON.stringify(record);
      tr.appendChild(n);
      tr.appendChild(v);
      rows.appendChild(tr);
    });
    status.textContent = records.length + ' record(s)';
  }

  function load() {
    return fetch('api/records')
      .then(function (r) { return r.json(); })
      .then(function (data) { render(data.records || []); });
  }

  if (refresh) {
    refresh.addEventListener('click', function () {
      refresh.disabled = true;
      status.textContent = 'Refreshing...';
      fetch('api/refresh', { method: 'POST' })
        .then(function (r) {
          if (!r.ok) return r.text().then(function (t) { throw new Error(t); });
          return r.json();
        })
        .then(load)
        .catch(function (err) { status.textContent = 'Error: ' + err.message; })
        .finally(function () { refresh.disabled = false; });
    });
  }

  load().catch(function (err) { status.textContent = 'Error: ' + err.message; });
})();
