package backend

const windowSyncCaptureScript = `(() => {
  if (globalThis.__lianxiangWindowSyncInstalled) return;
  if (typeof globalThis.__lianxiangWindowSyncEmit !== 'function') return;
  globalThis.__lianxiangWindowSyncInstalled = true;

  const emit = (payload) => {
    try { globalThis.__lianxiangWindowSyncEmit(JSON.stringify(payload)); } catch (_) {}
  };
  const selectorFor = (element) => {
    if (!(element instanceof Element)) return '';
    if (element.id) return '#' + CSS.escape(element.id);
    const parts = [];
    let current = element;
    while (current && current.nodeType === 1 && parts.length < 7) {
      let part = current.localName;
      if (!part) break;
      const testId = current.getAttribute('data-testid');
      if (testId) {
        parts.unshift('[data-testid="' + CSS.escape(testId) + '"]');
        break;
      }
      const name = current.getAttribute('name');
      if (name && /^(input|textarea|select|button)$/.test(part)) {
        parts.unshift(part + '[name="' + CSS.escape(name) + '"]');
        break;
      }
      const parent = current.parentElement;
      if (parent) {
        const peers = Array.from(parent.children).filter((item) => item.localName === current.localName);
        if (peers.length > 1) part += ':nth-of-type(' + (peers.indexOf(current) + 1) + ')';
      }
      parts.unshift(part);
      current = parent;
    }
    return parts.join(' > ');
  };

  document.addEventListener('click', (event) => {
    emit({
      type: 'click',
      selector: selectorFor(event.target),
      button: event.button === 1 ? 'middle' : event.button === 2 ? 'right' : 'left',
      clickCount: Math.max(1, event.detail || 1),
      xRatio: Math.max(0, Math.min(1, event.clientX / Math.max(1, innerWidth))),
      yRatio: Math.max(0, Math.min(1, event.clientY / Math.max(1, innerHeight)))
    });
  }, true);

  const lastInput = new WeakMap();
  const emitInput = (event) => {
    const target = event.target;
    if (!(target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement || target instanceof HTMLSelectElement || target?.isContentEditable)) return;
    const value = target.isContentEditable ? target.innerText : target.value;
    const checked = target instanceof HTMLInputElement ? target.checked : false;
    const snapshot = String(value) + '\u0000' + String(checked);
    if (lastInput.get(target) === snapshot) return;
    lastInput.set(target, snapshot);
    emit({
      type: 'input',
      selector: selectorFor(target),
      value: String(value ?? ''),
      checked,
      inputKind: target instanceof HTMLInputElement ? target.type : target.isContentEditable ? 'contenteditable' : target.localName
    });
  };
  document.addEventListener('input', emitInput, true);
  document.addEventListener('change', emitInput, true);

  document.addEventListener('keydown', (event) => {
    if (!['Enter', 'Escape', 'Tab'].includes(event.key)) return;
    emit({ type: 'key', selector: selectorFor(event.target), key: event.key, code: event.code || '' });
  }, true);

  let scrollTimer = 0;
  window.addEventListener('scroll', (event) => {
    if (scrollTimer) return;
    scrollTimer = window.setTimeout(() => {
      scrollTimer = 0;
      const rawTarget = event.target === document ? document.scrollingElement : event.target;
      const target = rawTarget instanceof Element ? rawTarget : document.scrollingElement;
      const isPage = target === document.documentElement || target === document.body || target === document.scrollingElement;
      const left = isPage ? scrollX : target.scrollLeft;
      const top = isPage ? scrollY : target.scrollTop;
      const width = isPage ? document.documentElement.scrollWidth - innerWidth : target.scrollWidth - target.clientWidth;
      const height = isPage ? document.documentElement.scrollHeight - innerHeight : target.scrollHeight - target.clientHeight;
      emit({
        type: 'scroll',
        selector: isPage ? '' : selectorFor(target),
        scrollX: left / Math.max(1, width),
        scrollY: top / Math.max(1, height)
      });
    }, 80);
  }, { passive: true, capture: true });
})();`

const windowSyncApplyInputExpression = `(() => {
  const payload = %s;
  const element = document.querySelector(payload.selector);
  if (!element) return false;
  if (element.isContentEditable) {
    element.innerText = payload.value || '';
  } else if (element instanceof HTMLInputElement && (element.type === 'checkbox' || element.type === 'radio')) {
    const descriptor = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'checked');
    if (descriptor && descriptor.set) descriptor.set.call(element, Boolean(payload.checked));
    else element.checked = Boolean(payload.checked);
  } else {
    const prototype = element instanceof HTMLTextAreaElement
      ? HTMLTextAreaElement.prototype
      : element instanceof HTMLSelectElement
        ? HTMLSelectElement.prototype
        : HTMLInputElement.prototype;
    const descriptor = Object.getOwnPropertyDescriptor(prototype, 'value');
    if (descriptor && descriptor.set) descriptor.set.call(element, payload.value || '');
    else element.value = payload.value || '';
  }
  element.dispatchEvent(new Event('input', { bubbles: true, composed: true }));
  element.dispatchEvent(new Event('change', { bubbles: true, composed: true }));
  return true;
})()`
