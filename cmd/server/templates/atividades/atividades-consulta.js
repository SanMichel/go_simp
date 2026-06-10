document.addEventListener("DOMContentLoaded", function() {
  var consultaMode = "codigo";

  function setConsultaMode(mode) {
    consultaMode = mode;
    var input = document.getElementById("consulta-input");
    var btnCodigo = document.getElementById("btn-consulta-mode-codigo");
    var btnDescricao = document.getElementById("btn-consulta-mode-descricao");
    if (mode === "codigo") {
      input.placeholder = "Escanear ou digitar EAN/DUN";
      input.type = "tel";
      input.inputMode = "none";
      btnCodigo.className = "btn btn-sm btn-primary";
      btnCodigo.style.cssText = "flex: 1; font-size: 0.75rem; padding: 0.35rem;";
      btnDescricao.className = "btn btn-sm";
      btnDescricao.style.cssText = "flex: 1; font-size: 0.75rem; padding: 0.35rem; background: #e2e8f0; color: #475569;";
    } else {
      input.placeholder = "Digitar descri\u00E7\u00E3o do produto";
      input.type = "text";
      input.inputMode = "text";
      btnDescricao.className = "btn btn-sm btn-primary";
      btnDescricao.style.cssText = "flex: 1; font-size: 0.75rem; padding: 0.35rem;";
      btnCodigo.className = "btn btn-sm";
      btnCodigo.style.cssText = "flex: 1; font-size: 0.75rem; padding: 0.35rem; background: #e2e8f0; color: #475569;";
    }
    var cr = document.getElementById("consulta-result");
    if (cr) cr.classList.add("hidden");
    var crl = document.getElementById("consulta-result-list");
    if (crl) {
      crl.classList.add("hidden");
      crl.innerHTML = "";
    }
    input.value = "";
    input.focus();
  }

  // Consulta mode toggle buttons
  var btnModoCodigo = document.getElementById("btn-consulta-mode-codigo");
  if (btnModoCodigo) {
    btnModoCodigo.addEventListener("click", function() { setConsultaMode("codigo"); });
  }
  var btnModoDescricao = document.getElementById("btn-consulta-mode-descricao");
  if (btnModoDescricao) {
    btnModoDescricao.addEventListener("click", function() { setConsultaMode("descricao"); });
  }

  // Consulta back button
  var btnConsultaBack = document.getElementById("btn-consulta-back");
  if (btnConsultaBack) {
    btnConsultaBack.addEventListener("click", function() {
      var backTo = state.previousScreen || (state.atividade ? "scanning" : "start");
      showScreen(backTo === "consulta" ? "start" : backTo);
    });
  }

  // Consulta form submit
  var formConsulta = document.getElementById("form-consulta");
  if (formConsulta) {
    formConsulta.addEventListener("submit", function(e) {
      e.preventDefault();
      var input = document.getElementById("consulta-input");
      var code = input.value.trim();
      if (!code) return;
      var empresa = null;
      var seqlocal = null;
      var lojaNome = "";
      if (state.previousScreen === "scanning" && state.atividade) {
        empresa = state.atividade.empresa;
        seqlocal = state.atividade.seqlocal;
        lojaNome = state.atividade.empresaNome || "Loja " + empresa;
      } else {
        var empSelect = document.getElementById("start-empresa");
        var locSelect = document.getElementById("start-local");
        empresa = empSelect ? empSelect.value : null;
        seqlocal = locSelect ? locSelect.value : null;
        if (empSelect && empSelect.selectedIndex >= 0) {
          var fullText = empSelect.options[empSelect.selectedIndex].text;
          lojaNome = fullText.split(" - ")[1] || fullText;
        }
      }
      if (!empresa || !seqlocal) {
        alert("Selecione uma empresa e local primeiro");
        return;
      }
      showLoader(true);
      var cr = document.getElementById("consulta-result");
      if (cr) cr.classList.add("hidden");
      var crl = document.getElementById("consulta-result-list");
      if (crl) {
        crl.classList.add("hidden");
        crl.innerHTML = "";
      }
      if (consultaMode === "descricao") {
        apiGet("/api/produtos/consulta?q=" + encodeURIComponent(code) + "&empresa=" + encodeURIComponent(empresa) + "&seqlocal=" + encodeURIComponent(seqlocal), function(ok, status, data) {
          showLoader(false);
          if (ok && Array.isArray(data) && data.length > 0) {
            playBeep("success");
            var empty = document.getElementById("consulta-empty");
            if (empty) empty.classList.add("hidden");
            var listEl = document.getElementById("consulta-result-list");
            var html = "";
            for (var i = 0; i < data.length; i++) {
              var p = data[i];
              var mdv = p.mdv != null ? Number(p.mdv).toFixed(2).replace(".", ",") : "\u2014";
              var preco = p.precoVenda ? "R$ " + Number(p.precoVenda).toFixed(2).replace(".", ",") : "N/A";
              var ultEntrada = p.dtaUltEntrada ? formatDate(p.dtaUltEntrada) : "N/A";
              var ultVenda = p.dtaUltVenda ? formatDate(p.dtaUltVenda) : "N/A";
              var codigosHtml = "";
              if (p.codigos) {
                var codigosArr = p.codigos.split("|");
                for (var ci = 0; ci < codigosArr.length; ci++) {
                  codigosHtml += '<span class="ean-badge">' + sanitizeHtml(codigosArr[ci]) + '</span> ';
                }
              }
              html += '<div class="card" style="margin-bottom: 0.5rem;">' +
                '<div style="display: flex; justify-content: space-between; align-items: flex-start;">' +
                  '<div style="flex: 1;">' +
                    '<div style="font-weight: 600; color: #4f46e5; font-size: 0.9rem;">' + sanitizeHtml(p.desccompleta || "Sem descri\u00E7\u00E3o") + '</div>' +
                    '<div style="font-size: 0.7rem; color: #64748b; margin-top: 2px;">SEQ ' + sanitizeHtml(String(p.seqproduto)) + ' \u00B7 ' + sanitizeHtml(p.marca || "\u2014") + '</div>' +
                  '</div>' +
                  '<span style="background: #f1f5f9; color: #475569; padding: 2px 6px; border-radius: 4px; font-size: 0.625rem; font-weight: 700; white-space: nowrap;">' + sanitizeHtml(lojaNome) + '</span>' +
                '</div>' +
                '<div class="grid-details" style="margin-top: 0.5rem;">' +
                  '<div class="detail-item"><span class="detail-label">ESTOQUE</span><span class="detail-value">' + sanitizeHtml(String(p.estoque)) + '</span></div>' +
                  '<div class="detail-item"><span class="detail-label">DIAS</span><span class="detail-value">' + (p.diasEstoque != null ? Number(p.diasEstoque).toFixed(1).replace(".", ",") : "\u2014") + '</span></div>' +
                  '<div class="detail-item"><span class="detail-label">MDV</span><span class="detail-value">' + mdv + '</span></div>' +
                  '<div class="detail-item"><span class="detail-label">PRE\u00c7O</span><span class="detail-value">' + preco + '</span></div>' +
                '</div>' +
                '<div class="mt-2 text-xs text-slate-500" style="border-top: 1px solid #e2e8f0; padding-top: 0.5rem;">' +
                  '<p>\u00DAltima Entrada: <span style="font-weight: 600;">' + sanitizeHtml(ultEntrada) + '</span></p>' +
                  '<p>\u00DAltima Venda: <span style="font-weight: 600;">' + sanitizeHtml(ultVenda) + '</span></p>' +
                  (codigosHtml ? '<p style="margin-top: 0.35rem; word-break: break-all;"><strong>C\u00D3DIGOS:</strong> ' + codigosHtml + '</p>' : "") +
                '</div>' +
              '</div>';
            }
            listEl.innerHTML = html;
            listEl.classList.remove("hidden");
          } else {
            playBeep("error");
            alert("Nenhum produto encontrado");
          }
          input.select();
        }, function() {
          showLoader(false);
          window.location.href = "/atividades/login";
        });
        return;
      }

      // Code search mode
      apiGet("/api/produtos/consulta/" + encodeURIComponent(code) + "?empresa=" + encodeURIComponent(empresa) + "&seqlocal=" + encodeURIComponent(seqlocal), function(ok, status, data) {
        showLoader(false);
        if (ok) {
          playBeep("success");
          var empty = document.getElementById("consulta-empty");
          if (empty) empty.classList.add("hidden");
          var resultEl = document.getElementById("consulta-result");
          if (resultEl) resultEl.classList.remove("hidden");
          function setVal(id, val) {
            var el = document.getElementById(id);
            if (el) el.innerText = String(val || "");
          }
          setVal("consulta-nome", data.desccompleta);
          setVal("consulta-loja-name", lojaNome);
          setVal("consulta-seq", data.seqproduto);
          setVal("consulta-marca", data.marca);
          setVal("consulta-estoque", data.estoque);
          setVal("consulta-dias", data.diasEstoque != null ? Number(data.diasEstoque).toFixed(1).replace(".", ",") : "\u2014");
          setVal("consulta-mdv", data.mdv != null ? Number(data.mdv).toFixed(2).replace(".", ",") : "\u2014");
          setVal("consulta-preco", data.precoVenda ? "R$ " + Number(data.precoVenda).toFixed(2).replace(".", ",") : "N/A");
          setVal("consulta-entrada", data.dtaUltEntrada ? formatDate(data.dtaUltEntrada) : "N/A");
          setVal("consulta-venda", data.dtaUltVenda ? formatDate(data.dtaUltVenda) : "N/A");
          var codigosEl = document.getElementById("consulta-codigos");
          if (codigosEl && data.codigos) {
            var codArr = data.codigos.split("|");
            var codHtml = "";
            for (var ci2 = 0; ci2 < codArr.length; ci2++) {
              codHtml += '<span class="ean-badge">' + sanitizeHtml(codArr[ci2]) + '</span> ';
            }
            codigosEl.innerHTML = sanitizeHtml(codHtml);
          }
          input.select();
        } else {
          playBeep("error");
          alert("Produto n\u00E3o encontrado");
          input.select();
        }
      }, function() {
        showLoader(false);
        window.location.href = "/atividades/login";
      });
    });
  }
});
