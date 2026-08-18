const status = document.getElementById("status");
const tree = document.getElementById("tree");

function refresh() {
  chrome.devtools.inspectedWindow.eval(
    "window.__SWC_DEVTOOLS__ ? JSON.stringify(window.__SWC_DEVTOOLS__.components.map(c => ({ id: c.id, name: c.name }))) : '[]'",
    (result, err) => {
      if (err || !result) {
        if (status) status.textContent = "SWC not detected on this page.";
        return;
      }
      const comps = JSON.parse(result as string) as Array<{ id: number; name: string }>;
      if (status) status.textContent = `${comps.length} component(s)`;
      if (tree) {
        tree.innerHTML = comps.map((c) => `<li>${c.name} #${c.id}</li>`).join("");
      }
    },
  );
}

refresh();
setInterval(refresh, 2000);
