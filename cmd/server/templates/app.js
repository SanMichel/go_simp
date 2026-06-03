// src/client/shared/api.ts
async function apiCall(endpoint, options = {}, onUnauthorized) {
  const headers = {
    "Content-Type": "application/json"
  };
  async function execute() {
    return fetch(endpoint, {
      ...options,
      headers: { ...headers, ...options.headers },
      credentials: "include"
    });
  }
  try {
    let res = await execute();
    if (res.status === 401) {
      const refreshRes = await fetch("/api/auth/refresh", {
        method: "POST",
        credentials: "include"
      });
      if (refreshRes.ok) {
        res = await execute();
      } else {
        if (onUnauthorized)
          onUnauthorized();
        throw new Error("Unauthorized");
      }
    }
    let data;
    const contentType = res.headers.get("content-type");
    if (contentType?.includes("application/json")) {
      data = await res.json();
    } else {
      const text = await res.text();
      data = { error: text || `Error ${res.status}: ${res.statusText}` };
    }
    return { ok: res.ok, status: res.status, data };
  } catch (e) {
    if (e instanceof Error && e.message === "Unauthorized") {
      return { ok: false, status: 401, data: { error: "Unauthorized" } };
    }
    console.error(`API Call failed: ${endpoint}`, e);
    return {
      ok: false,
      status: 0,
      data: {
        error: e instanceof Error ? e.message : "Erro de conexão"
      }
    };
  }
}

// node_modules/dompurify/dist/purify.es.mjs
/*! @license DOMPurify 3.4.2 | (c) Cure53 and other contributors | Released under the Apache license 2.0 and Mozilla Public License 2.0 | github.com/cure53/DOMPurify/blob/3.4.2/LICENSE */
var {
  entries,
  setPrototypeOf,
  isFrozen,
  getPrototypeOf,
  getOwnPropertyDescriptor
} = Object;
var {
  freeze,
  seal,
  create
} = Object;
var {
  apply,
  construct
} = typeof Reflect !== "undefined" && Reflect;
if (!freeze) {
  freeze = function freeze2(x) {
    return x;
  };
}
if (!seal) {
  seal = function seal2(x) {
    return x;
  };
}
if (!apply) {
  apply = function apply2(func, thisArg) {
    for (var _len = arguments.length, args = new Array(_len > 2 ? _len - 2 : 0), _key = 2;_key < _len; _key++) {
      args[_key - 2] = arguments[_key];
    }
    return func.apply(thisArg, args);
  };
}
if (!construct) {
  construct = function construct2(Func) {
    for (var _len2 = arguments.length, args = new Array(_len2 > 1 ? _len2 - 1 : 0), _key2 = 1;_key2 < _len2; _key2++) {
      args[_key2 - 1] = arguments[_key2];
    }
    return new Func(...args);
  };
}
var arrayForEach = unapply(Array.prototype.forEach);
var arrayLastIndexOf = unapply(Array.prototype.lastIndexOf);
var arrayPop = unapply(Array.prototype.pop);
var arrayPush = unapply(Array.prototype.push);
var arraySplice = unapply(Array.prototype.splice);
var arrayIsArray = Array.isArray;
var stringToLowerCase = unapply(String.prototype.toLowerCase);
var stringToString = unapply(String.prototype.toString);
var stringMatch = unapply(String.prototype.match);
var stringReplace = unapply(String.prototype.replace);
var stringIndexOf = unapply(String.prototype.indexOf);
var stringTrim = unapply(String.prototype.trim);
var numberToString = unapply(Number.prototype.toString);
var booleanToString = unapply(Boolean.prototype.toString);
var bigintToString = typeof BigInt === "undefined" ? null : unapply(BigInt.prototype.toString);
var symbolToString = typeof Symbol === "undefined" ? null : unapply(Symbol.prototype.toString);
var objectHasOwnProperty = unapply(Object.prototype.hasOwnProperty);
var objectToString = unapply(Object.prototype.toString);
var regExpTest = unapply(RegExp.prototype.test);
var typeErrorCreate = unconstruct(TypeError);
function unapply(func) {
  return function(thisArg) {
    if (thisArg instanceof RegExp) {
      thisArg.lastIndex = 0;
    }
    for (var _len3 = arguments.length, args = new Array(_len3 > 1 ? _len3 - 1 : 0), _key3 = 1;_key3 < _len3; _key3++) {
      args[_key3 - 1] = arguments[_key3];
    }
    return apply(func, thisArg, args);
  };
}
function unconstruct(Func) {
  return function() {
    for (var _len4 = arguments.length, args = new Array(_len4), _key4 = 0;_key4 < _len4; _key4++) {
      args[_key4] = arguments[_key4];
    }
    return construct(Func, args);
  };
}
function addToSet(set, array) {
  let transformCaseFunc = arguments.length > 2 && arguments[2] !== undefined ? arguments[2] : stringToLowerCase;
  if (setPrototypeOf) {
    setPrototypeOf(set, null);
  }
  if (!arrayIsArray(array)) {
    return set;
  }
  let l = array.length;
  while (l--) {
    let element = array[l];
    if (typeof element === "string") {
      const lcElement = transformCaseFunc(element);
      if (lcElement !== element) {
        if (!isFrozen(array)) {
          array[l] = lcElement;
        }
        element = lcElement;
      }
    }
    set[element] = true;
  }
  return set;
}
function cleanArray(array) {
  for (let index = 0;index < array.length; index++) {
    const isPropertyExist = objectHasOwnProperty(array, index);
    if (!isPropertyExist) {
      array[index] = null;
    }
  }
  return array;
}
function clone(object) {
  const newObject = create(null);
  for (const [property, value] of entries(object)) {
    const isPropertyExist = objectHasOwnProperty(object, property);
    if (isPropertyExist) {
      if (arrayIsArray(value)) {
        newObject[property] = cleanArray(value);
      } else if (value && typeof value === "object" && value.constructor === Object) {
        newObject[property] = clone(value);
      } else {
        newObject[property] = value;
      }
    }
  }
  return newObject;
}
function stringifyValue(value) {
  switch (typeof value) {
    case "string": {
      return value;
    }
    case "number": {
      return numberToString(value);
    }
    case "boolean": {
      return booleanToString(value);
    }
    case "bigint": {
      return bigintToString ? bigintToString(value) : "0";
    }
    case "symbol": {
      return symbolToString ? symbolToString(value) : "Symbol()";
    }
    case "undefined": {
      return objectToString(value);
    }
    case "function":
    case "object": {
      if (value === null) {
        return objectToString(value);
      }
      const valueAsRecord = value;
      const valueToString = lookupGetter(valueAsRecord, "toString");
      if (typeof valueToString === "function") {
        const stringified = valueToString(valueAsRecord);
        return typeof stringified === "string" ? stringified : objectToString(stringified);
      }
      return objectToString(value);
    }
    default: {
      return objectToString(value);
    }
  }
}
function lookupGetter(object, prop) {
  while (object !== null) {
    const desc = getOwnPropertyDescriptor(object, prop);
    if (desc) {
      if (desc.get) {
        return unapply(desc.get);
      }
      if (typeof desc.value === "function") {
        return unapply(desc.value);
      }
    }
    object = getPrototypeOf(object);
  }
  function fallbackValue() {
    return null;
  }
  return fallbackValue;
}
function isRegex(value) {
  try {
    regExpTest(value, "");
    return true;
  } catch (_unused) {
    return false;
  }
}
var html$1 = freeze(["a", "abbr", "acronym", "address", "area", "article", "aside", "audio", "b", "bdi", "bdo", "big", "blink", "blockquote", "body", "br", "button", "canvas", "caption", "center", "cite", "code", "col", "colgroup", "content", "data", "datalist", "dd", "decorator", "del", "details", "dfn", "dialog", "dir", "div", "dl", "dt", "element", "em", "fieldset", "figcaption", "figure", "font", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6", "head", "header", "hgroup", "hr", "html", "i", "img", "input", "ins", "kbd", "label", "legend", "li", "main", "map", "mark", "marquee", "menu", "menuitem", "meter", "nav", "nobr", "ol", "optgroup", "option", "output", "p", "picture", "pre", "progress", "q", "rp", "rt", "ruby", "s", "samp", "search", "section", "select", "shadow", "slot", "small", "source", "spacer", "span", "strike", "strong", "style", "sub", "summary", "sup", "table", "tbody", "td", "template", "textarea", "tfoot", "th", "thead", "time", "tr", "track", "tt", "u", "ul", "var", "video", "wbr"]);
var svg$1 = freeze(["svg", "a", "altglyph", "altglyphdef", "altglyphitem", "animatecolor", "animatemotion", "animatetransform", "circle", "clippath", "defs", "desc", "ellipse", "enterkeyhint", "exportparts", "filter", "font", "g", "glyph", "glyphref", "hkern", "image", "inputmode", "line", "lineargradient", "marker", "mask", "metadata", "mpath", "part", "path", "pattern", "polygon", "polyline", "radialgradient", "rect", "stop", "style", "switch", "symbol", "text", "textpath", "title", "tref", "tspan", "view", "vkern"]);
var svgFilters = freeze(["feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDistantLight", "feDropShadow", "feFlood", "feFuncA", "feFuncB", "feFuncG", "feFuncR", "feGaussianBlur", "feImage", "feMerge", "feMergeNode", "feMorphology", "feOffset", "fePointLight", "feSpecularLighting", "feSpotLight", "feTile", "feTurbulence"]);
var svgDisallowed = freeze(["animate", "color-profile", "cursor", "discard", "font-face", "font-face-format", "font-face-name", "font-face-src", "font-face-uri", "foreignobject", "hatch", "hatchpath", "mesh", "meshgradient", "meshpatch", "meshrow", "missing-glyph", "script", "set", "solidcolor", "unknown", "use"]);
var mathMl$1 = freeze(["math", "menclose", "merror", "mfenced", "mfrac", "mglyph", "mi", "mlabeledtr", "mmultiscripts", "mn", "mo", "mover", "mpadded", "mphantom", "mroot", "mrow", "ms", "mspace", "msqrt", "mstyle", "msub", "msup", "msubsup", "mtable", "mtd", "mtext", "mtr", "munder", "munderover", "mprescripts"]);
var mathMlDisallowed = freeze(["maction", "maligngroup", "malignmark", "mlongdiv", "mscarries", "mscarry", "msgroup", "mstack", "msline", "msrow", "semantics", "annotation", "annotation-xml", "mprescripts", "none"]);
var text = freeze(["#text"]);
var html = freeze(["accept", "action", "align", "alt", "autocapitalize", "autocomplete", "autopictureinpicture", "autoplay", "background", "bgcolor", "border", "capture", "cellpadding", "cellspacing", "checked", "cite", "class", "clear", "color", "cols", "colspan", "controls", "controlslist", "coords", "crossorigin", "datetime", "decoding", "default", "dir", "disabled", "disablepictureinpicture", "disableremoteplayback", "download", "draggable", "enctype", "enterkeyhint", "exportparts", "face", "for", "headers", "height", "hidden", "high", "href", "hreflang", "id", "inert", "inputmode", "integrity", "ismap", "kind", "label", "lang", "list", "loading", "loop", "low", "max", "maxlength", "media", "method", "min", "minlength", "multiple", "muted", "name", "nonce", "noshade", "novalidate", "nowrap", "open", "optimum", "part", "pattern", "placeholder", "playsinline", "popover", "popovertarget", "popovertargetaction", "poster", "preload", "pubdate", "radiogroup", "readonly", "rel", "required", "rev", "reversed", "role", "rows", "rowspan", "spellcheck", "scope", "selected", "shape", "size", "sizes", "slot", "span", "srclang", "start", "src", "srcset", "step", "style", "summary", "tabindex", "title", "translate", "type", "usemap", "valign", "value", "width", "wrap", "xmlns"]);
var svg = freeze(["accent-height", "accumulate", "additive", "alignment-baseline", "amplitude", "ascent", "attributename", "attributetype", "azimuth", "basefrequency", "baseline-shift", "begin", "bias", "by", "class", "clip", "clippathunits", "clip-path", "clip-rule", "color", "color-interpolation", "color-interpolation-filters", "color-profile", "color-rendering", "cx", "cy", "d", "dx", "dy", "diffuseconstant", "direction", "display", "divisor", "dur", "edgemode", "elevation", "end", "exponent", "fill", "fill-opacity", "fill-rule", "filter", "filterunits", "flood-color", "flood-opacity", "font-family", "font-size", "font-size-adjust", "font-stretch", "font-style", "font-variant", "font-weight", "fx", "fy", "g1", "g2", "glyph-name", "glyphref", "gradientunits", "gradienttransform", "height", "href", "id", "image-rendering", "in", "in2", "intercept", "k", "k1", "k2", "k3", "k4", "kerning", "keypoints", "keysplines", "keytimes", "lang", "lengthadjust", "letter-spacing", "kernelmatrix", "kernelunitlength", "lighting-color", "local", "marker-end", "marker-mid", "marker-start", "markerheight", "markerunits", "markerwidth", "maskcontentunits", "maskunits", "max", "mask", "mask-type", "media", "method", "mode", "min", "name", "numoctaves", "offset", "operator", "opacity", "order", "orient", "orientation", "origin", "overflow", "paint-order", "path", "pathlength", "patterncontentunits", "patterntransform", "patternunits", "points", "preservealpha", "preserveaspectratio", "primitiveunits", "r", "rx", "ry", "radius", "refx", "refy", "repeatcount", "repeatdur", "restart", "result", "rotate", "scale", "seed", "shape-rendering", "slope", "specularconstant", "specularexponent", "spreadmethod", "startoffset", "stddeviation", "stitchtiles", "stop-color", "stop-opacity", "stroke-dasharray", "stroke-dashoffset", "stroke-linecap", "stroke-linejoin", "stroke-miterlimit", "stroke-opacity", "stroke", "stroke-width", "style", "surfacescale", "systemlanguage", "tabindex", "tablevalues", "targetx", "targety", "transform", "transform-origin", "text-anchor", "text-decoration", "text-rendering", "textlength", "type", "u1", "u2", "unicode", "values", "viewbox", "visibility", "version", "vert-adv-y", "vert-origin-x", "vert-origin-y", "width", "word-spacing", "wrap", "writing-mode", "xchannelselector", "ychannelselector", "x", "x1", "x2", "xmlns", "y", "y1", "y2", "z", "zoomandpan"]);
var mathMl = freeze(["accent", "accentunder", "align", "bevelled", "close", "columnalign", "columnlines", "columnspacing", "columnspan", "denomalign", "depth", "dir", "display", "displaystyle", "encoding", "fence", "frame", "height", "href", "id", "largeop", "length", "linethickness", "lquote", "lspace", "mathbackground", "mathcolor", "mathsize", "mathvariant", "maxsize", "minsize", "movablelimits", "notation", "numalign", "open", "rowalign", "rowlines", "rowspacing", "rowspan", "rspace", "rquote", "scriptlevel", "scriptminsize", "scriptsizemultiplier", "selection", "separator", "separators", "stretchy", "subscriptshift", "supscriptshift", "symmetric", "voffset", "width", "xmlns"]);
var xml = freeze(["xlink:href", "xml:id", "xlink:title", "xml:space", "xmlns:xlink"]);
var MUSTACHE_EXPR = seal(/\{\{[\w\W]*|[\w\W]*\}\}/gm);
var ERB_EXPR = seal(/<%[\w\W]*|[\w\W]*%>/gm);
var TMPLIT_EXPR = seal(/\$\{[\w\W]*/gm);
var DATA_ATTR = seal(/^data-[\-\w.\u00B7-\uFFFF]+$/);
var ARIA_ATTR = seal(/^aria-[\-\w]+$/);
var IS_ALLOWED_URI = seal(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i);
var IS_SCRIPT_OR_DATA = seal(/^(?:\w+script|data):/i);
var ATTR_WHITESPACE = seal(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g);
var DOCTYPE_NAME = seal(/^html$/i);
var CUSTOM_ELEMENT = seal(/^[a-z][.\w]*(-[.\w]+)+$/i);
var EXPRESSIONS = /* @__PURE__ */ Object.freeze({
  __proto__: null,
  ARIA_ATTR,
  ATTR_WHITESPACE,
  CUSTOM_ELEMENT,
  DATA_ATTR,
  DOCTYPE_NAME,
  ERB_EXPR,
  IS_ALLOWED_URI,
  IS_SCRIPT_OR_DATA,
  MUSTACHE_EXPR,
  TMPLIT_EXPR
});
var NODE_TYPE = {
  element: 1,
  text: 3,
  progressingInstruction: 7,
  comment: 8,
  document: 9
};
var getGlobal = function getGlobal2() {
  return typeof window === "undefined" ? null : window;
};
var _createTrustedTypesPolicy = function _createTrustedTypesPolicy2(trustedTypes, purifyHostElement) {
  if (typeof trustedTypes !== "object" || typeof trustedTypes.createPolicy !== "function") {
    return null;
  }
  let suffix = null;
  const ATTR_NAME = "data-tt-policy-suffix";
  if (purifyHostElement && purifyHostElement.hasAttribute(ATTR_NAME)) {
    suffix = purifyHostElement.getAttribute(ATTR_NAME);
  }
  const policyName = "dompurify" + (suffix ? "#" + suffix : "");
  try {
    return trustedTypes.createPolicy(policyName, {
      createHTML(html2) {
        return html2;
      },
      createScriptURL(scriptUrl) {
        return scriptUrl;
      }
    });
  } catch (_) {
    console.warn("TrustedTypes policy " + policyName + " could not be created.");
    return null;
  }
};
var _createHooksMap = function _createHooksMap2() {
  return {
    afterSanitizeAttributes: [],
    afterSanitizeElements: [],
    afterSanitizeShadowDOM: [],
    beforeSanitizeAttributes: [],
    beforeSanitizeElements: [],
    beforeSanitizeShadowDOM: [],
    uponSanitizeAttribute: [],
    uponSanitizeElement: [],
    uponSanitizeShadowNode: []
  };
};
function createDOMPurify() {
  let window2 = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : getGlobal();
  const DOMPurify = (root) => createDOMPurify(root);
  DOMPurify.version = "3.4.2";
  DOMPurify.removed = [];
  if (!window2 || !window2.document || window2.document.nodeType !== NODE_TYPE.document || !window2.Element) {
    DOMPurify.isSupported = false;
    return DOMPurify;
  }
  let {
    document: document2
  } = window2;
  const originalDocument = document2;
  const currentScript = originalDocument.currentScript;
  const {
    DocumentFragment,
    HTMLTemplateElement,
    Node,
    Element,
    NodeFilter,
    NamedNodeMap = window2.NamedNodeMap || window2.MozNamedAttrMap,
    HTMLFormElement,
    DOMParser,
    trustedTypes
  } = window2;
  const ElementPrototype = Element.prototype;
  const cloneNode = lookupGetter(ElementPrototype, "cloneNode");
  const remove = lookupGetter(ElementPrototype, "remove");
  const getNextSibling = lookupGetter(ElementPrototype, "nextSibling");
  const getChildNodes = lookupGetter(ElementPrototype, "childNodes");
  const getParentNode = lookupGetter(ElementPrototype, "parentNode");
  if (typeof HTMLTemplateElement === "function") {
    const template = document2.createElement("template");
    if (template.content && template.content.ownerDocument) {
      document2 = template.content.ownerDocument;
    }
  }
  let trustedTypesPolicy;
  let emptyHTML = "";
  const {
    implementation,
    createNodeIterator,
    createDocumentFragment,
    getElementsByTagName
  } = document2;
  const {
    importNode
  } = originalDocument;
  let hooks = _createHooksMap();
  DOMPurify.isSupported = typeof entries === "function" && typeof getParentNode === "function" && implementation && implementation.createHTMLDocument !== undefined;
  const {
    MUSTACHE_EXPR: MUSTACHE_EXPR2,
    ERB_EXPR: ERB_EXPR2,
    TMPLIT_EXPR: TMPLIT_EXPR2,
    DATA_ATTR: DATA_ATTR2,
    ARIA_ATTR: ARIA_ATTR2,
    IS_SCRIPT_OR_DATA: IS_SCRIPT_OR_DATA2,
    ATTR_WHITESPACE: ATTR_WHITESPACE2,
    CUSTOM_ELEMENT: CUSTOM_ELEMENT2
  } = EXPRESSIONS;
  let {
    IS_ALLOWED_URI: IS_ALLOWED_URI$1
  } = EXPRESSIONS;
  let ALLOWED_TAGS = null;
  const DEFAULT_ALLOWED_TAGS = addToSet({}, [...html$1, ...svg$1, ...svgFilters, ...mathMl$1, ...text]);
  let ALLOWED_ATTR = null;
  const DEFAULT_ALLOWED_ATTR = addToSet({}, [...html, ...svg, ...mathMl, ...xml]);
  let CUSTOM_ELEMENT_HANDLING = Object.seal(create(null, {
    tagNameCheck: {
      writable: true,
      configurable: false,
      enumerable: true,
      value: null
    },
    attributeNameCheck: {
      writable: true,
      configurable: false,
      enumerable: true,
      value: null
    },
    allowCustomizedBuiltInElements: {
      writable: true,
      configurable: false,
      enumerable: true,
      value: false
    }
  }));
  let FORBID_TAGS = null;
  let FORBID_ATTR = null;
  const EXTRA_ELEMENT_HANDLING = Object.seal(create(null, {
    tagCheck: {
      writable: true,
      configurable: false,
      enumerable: true,
      value: null
    },
    attributeCheck: {
      writable: true,
      configurable: false,
      enumerable: true,
      value: null
    }
  }));
  let ALLOW_ARIA_ATTR = true;
  let ALLOW_DATA_ATTR = true;
  let ALLOW_UNKNOWN_PROTOCOLS = false;
  let ALLOW_SELF_CLOSE_IN_ATTR = true;
  let SAFE_FOR_TEMPLATES = false;
  let SAFE_FOR_XML = true;
  let WHOLE_DOCUMENT = false;
  let SET_CONFIG = false;
  let FORCE_BODY = false;
  let RETURN_DOM = false;
  let RETURN_DOM_FRAGMENT = false;
  let RETURN_TRUSTED_TYPE = false;
  let SANITIZE_DOM = true;
  let SANITIZE_NAMED_PROPS = false;
  const SANITIZE_NAMED_PROPS_PREFIX = "user-content-";
  let KEEP_CONTENT = true;
  let IN_PLACE = false;
  let USE_PROFILES = {};
  let FORBID_CONTENTS = null;
  const DEFAULT_FORBID_CONTENTS = addToSet({}, ["annotation-xml", "audio", "colgroup", "desc", "foreignobject", "head", "iframe", "math", "mi", "mn", "mo", "ms", "mtext", "noembed", "noframes", "noscript", "plaintext", "script", "style", "svg", "template", "thead", "title", "video", "xmp"]);
  let DATA_URI_TAGS = null;
  const DEFAULT_DATA_URI_TAGS = addToSet({}, ["audio", "video", "img", "source", "image", "track"]);
  let URI_SAFE_ATTRIBUTES = null;
  const DEFAULT_URI_SAFE_ATTRIBUTES = addToSet({}, ["alt", "class", "for", "id", "label", "name", "pattern", "placeholder", "role", "summary", "title", "value", "style", "xmlns"]);
  const MATHML_NAMESPACE = "http://www.w3.org/1998/Math/MathML";
  const SVG_NAMESPACE = "http://www.w3.org/2000/svg";
  const HTML_NAMESPACE = "http://www.w3.org/1999/xhtml";
  let NAMESPACE = HTML_NAMESPACE;
  let IS_EMPTY_INPUT = false;
  let ALLOWED_NAMESPACES = null;
  const DEFAULT_ALLOWED_NAMESPACES = addToSet({}, [MATHML_NAMESPACE, SVG_NAMESPACE, HTML_NAMESPACE], stringToString);
  let MATHML_TEXT_INTEGRATION_POINTS = addToSet({}, ["mi", "mo", "mn", "ms", "mtext"]);
  let HTML_INTEGRATION_POINTS = addToSet({}, ["annotation-xml"]);
  const COMMON_SVG_AND_HTML_ELEMENTS = addToSet({}, ["title", "style", "font", "a", "script"]);
  let PARSER_MEDIA_TYPE = null;
  const SUPPORTED_PARSER_MEDIA_TYPES = ["application/xhtml+xml", "text/html"];
  const DEFAULT_PARSER_MEDIA_TYPE = "text/html";
  let transformCaseFunc = null;
  let CONFIG = null;
  const formElement = document2.createElement("form");
  const isRegexOrFunction = function isRegexOrFunction2(testValue) {
    return testValue instanceof RegExp || testValue instanceof Function;
  };
  const _parseConfig = function _parseConfig2() {
    let cfg = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};
    if (CONFIG && CONFIG === cfg) {
      return;
    }
    if (!cfg || typeof cfg !== "object") {
      cfg = {};
    }
    cfg = clone(cfg);
    PARSER_MEDIA_TYPE = SUPPORTED_PARSER_MEDIA_TYPES.indexOf(cfg.PARSER_MEDIA_TYPE) === -1 ? DEFAULT_PARSER_MEDIA_TYPE : cfg.PARSER_MEDIA_TYPE;
    transformCaseFunc = PARSER_MEDIA_TYPE === "application/xhtml+xml" ? stringToString : stringToLowerCase;
    ALLOWED_TAGS = objectHasOwnProperty(cfg, "ALLOWED_TAGS") && arrayIsArray(cfg.ALLOWED_TAGS) ? addToSet({}, cfg.ALLOWED_TAGS, transformCaseFunc) : DEFAULT_ALLOWED_TAGS;
    ALLOWED_ATTR = objectHasOwnProperty(cfg, "ALLOWED_ATTR") && arrayIsArray(cfg.ALLOWED_ATTR) ? addToSet({}, cfg.ALLOWED_ATTR, transformCaseFunc) : DEFAULT_ALLOWED_ATTR;
    ALLOWED_NAMESPACES = objectHasOwnProperty(cfg, "ALLOWED_NAMESPACES") && arrayIsArray(cfg.ALLOWED_NAMESPACES) ? addToSet({}, cfg.ALLOWED_NAMESPACES, stringToString) : DEFAULT_ALLOWED_NAMESPACES;
    URI_SAFE_ATTRIBUTES = objectHasOwnProperty(cfg, "ADD_URI_SAFE_ATTR") && arrayIsArray(cfg.ADD_URI_SAFE_ATTR) ? addToSet(clone(DEFAULT_URI_SAFE_ATTRIBUTES), cfg.ADD_URI_SAFE_ATTR, transformCaseFunc) : DEFAULT_URI_SAFE_ATTRIBUTES;
    DATA_URI_TAGS = objectHasOwnProperty(cfg, "ADD_DATA_URI_TAGS") && arrayIsArray(cfg.ADD_DATA_URI_TAGS) ? addToSet(clone(DEFAULT_DATA_URI_TAGS), cfg.ADD_DATA_URI_TAGS, transformCaseFunc) : DEFAULT_DATA_URI_TAGS;
    FORBID_CONTENTS = objectHasOwnProperty(cfg, "FORBID_CONTENTS") && arrayIsArray(cfg.FORBID_CONTENTS) ? addToSet({}, cfg.FORBID_CONTENTS, transformCaseFunc) : DEFAULT_FORBID_CONTENTS;
    FORBID_TAGS = objectHasOwnProperty(cfg, "FORBID_TAGS") && arrayIsArray(cfg.FORBID_TAGS) ? addToSet({}, cfg.FORBID_TAGS, transformCaseFunc) : clone({});
    FORBID_ATTR = objectHasOwnProperty(cfg, "FORBID_ATTR") && arrayIsArray(cfg.FORBID_ATTR) ? addToSet({}, cfg.FORBID_ATTR, transformCaseFunc) : clone({});
    USE_PROFILES = objectHasOwnProperty(cfg, "USE_PROFILES") ? cfg.USE_PROFILES && typeof cfg.USE_PROFILES === "object" ? clone(cfg.USE_PROFILES) : cfg.USE_PROFILES : false;
    ALLOW_ARIA_ATTR = cfg.ALLOW_ARIA_ATTR !== false;
    ALLOW_DATA_ATTR = cfg.ALLOW_DATA_ATTR !== false;
    ALLOW_UNKNOWN_PROTOCOLS = cfg.ALLOW_UNKNOWN_PROTOCOLS || false;
    ALLOW_SELF_CLOSE_IN_ATTR = cfg.ALLOW_SELF_CLOSE_IN_ATTR !== false;
    SAFE_FOR_TEMPLATES = cfg.SAFE_FOR_TEMPLATES || false;
    SAFE_FOR_XML = cfg.SAFE_FOR_XML !== false;
    WHOLE_DOCUMENT = cfg.WHOLE_DOCUMENT || false;
    RETURN_DOM = cfg.RETURN_DOM || false;
    RETURN_DOM_FRAGMENT = cfg.RETURN_DOM_FRAGMENT || false;
    RETURN_TRUSTED_TYPE = cfg.RETURN_TRUSTED_TYPE || false;
    FORCE_BODY = cfg.FORCE_BODY || false;
    SANITIZE_DOM = cfg.SANITIZE_DOM !== false;
    SANITIZE_NAMED_PROPS = cfg.SANITIZE_NAMED_PROPS || false;
    KEEP_CONTENT = cfg.KEEP_CONTENT !== false;
    IN_PLACE = cfg.IN_PLACE || false;
    IS_ALLOWED_URI$1 = isRegex(cfg.ALLOWED_URI_REGEXP) ? cfg.ALLOWED_URI_REGEXP : IS_ALLOWED_URI;
    NAMESPACE = typeof cfg.NAMESPACE === "string" ? cfg.NAMESPACE : HTML_NAMESPACE;
    MATHML_TEXT_INTEGRATION_POINTS = objectHasOwnProperty(cfg, "MATHML_TEXT_INTEGRATION_POINTS") && cfg.MATHML_TEXT_INTEGRATION_POINTS && typeof cfg.MATHML_TEXT_INTEGRATION_POINTS === "object" ? clone(cfg.MATHML_TEXT_INTEGRATION_POINTS) : addToSet({}, ["mi", "mo", "mn", "ms", "mtext"]);
    HTML_INTEGRATION_POINTS = objectHasOwnProperty(cfg, "HTML_INTEGRATION_POINTS") && cfg.HTML_INTEGRATION_POINTS && typeof cfg.HTML_INTEGRATION_POINTS === "object" ? clone(cfg.HTML_INTEGRATION_POINTS) : addToSet({}, ["annotation-xml"]);
    const customElementHandling = objectHasOwnProperty(cfg, "CUSTOM_ELEMENT_HANDLING") && cfg.CUSTOM_ELEMENT_HANDLING && typeof cfg.CUSTOM_ELEMENT_HANDLING === "object" ? clone(cfg.CUSTOM_ELEMENT_HANDLING) : create(null);
    CUSTOM_ELEMENT_HANDLING = create(null);
    if (objectHasOwnProperty(customElementHandling, "tagNameCheck") && isRegexOrFunction(customElementHandling.tagNameCheck)) {
      CUSTOM_ELEMENT_HANDLING.tagNameCheck = customElementHandling.tagNameCheck;
    }
    if (objectHasOwnProperty(customElementHandling, "attributeNameCheck") && isRegexOrFunction(customElementHandling.attributeNameCheck)) {
      CUSTOM_ELEMENT_HANDLING.attributeNameCheck = customElementHandling.attributeNameCheck;
    }
    if (objectHasOwnProperty(customElementHandling, "allowCustomizedBuiltInElements") && typeof customElementHandling.allowCustomizedBuiltInElements === "boolean") {
      CUSTOM_ELEMENT_HANDLING.allowCustomizedBuiltInElements = customElementHandling.allowCustomizedBuiltInElements;
    }
    if (SAFE_FOR_TEMPLATES) {
      ALLOW_DATA_ATTR = false;
    }
    if (RETURN_DOM_FRAGMENT) {
      RETURN_DOM = true;
    }
    if (USE_PROFILES) {
      ALLOWED_TAGS = addToSet({}, text);
      ALLOWED_ATTR = create(null);
      if (USE_PROFILES.html === true) {
        addToSet(ALLOWED_TAGS, html$1);
        addToSet(ALLOWED_ATTR, html);
      }
      if (USE_PROFILES.svg === true) {
        addToSet(ALLOWED_TAGS, svg$1);
        addToSet(ALLOWED_ATTR, svg);
        addToSet(ALLOWED_ATTR, xml);
      }
      if (USE_PROFILES.svgFilters === true) {
        addToSet(ALLOWED_TAGS, svgFilters);
        addToSet(ALLOWED_ATTR, svg);
        addToSet(ALLOWED_ATTR, xml);
      }
      if (USE_PROFILES.mathMl === true) {
        addToSet(ALLOWED_TAGS, mathMl$1);
        addToSet(ALLOWED_ATTR, mathMl);
        addToSet(ALLOWED_ATTR, xml);
      }
    }
    EXTRA_ELEMENT_HANDLING.tagCheck = null;
    EXTRA_ELEMENT_HANDLING.attributeCheck = null;
    if (objectHasOwnProperty(cfg, "ADD_TAGS")) {
      if (typeof cfg.ADD_TAGS === "function") {
        EXTRA_ELEMENT_HANDLING.tagCheck = cfg.ADD_TAGS;
      } else if (arrayIsArray(cfg.ADD_TAGS)) {
        if (ALLOWED_TAGS === DEFAULT_ALLOWED_TAGS) {
          ALLOWED_TAGS = clone(ALLOWED_TAGS);
        }
        addToSet(ALLOWED_TAGS, cfg.ADD_TAGS, transformCaseFunc);
      }
    }
    if (objectHasOwnProperty(cfg, "ADD_ATTR")) {
      if (typeof cfg.ADD_ATTR === "function") {
        EXTRA_ELEMENT_HANDLING.attributeCheck = cfg.ADD_ATTR;
      } else if (arrayIsArray(cfg.ADD_ATTR)) {
        if (ALLOWED_ATTR === DEFAULT_ALLOWED_ATTR) {
          ALLOWED_ATTR = clone(ALLOWED_ATTR);
        }
        addToSet(ALLOWED_ATTR, cfg.ADD_ATTR, transformCaseFunc);
      }
    }
    if (objectHasOwnProperty(cfg, "ADD_URI_SAFE_ATTR") && arrayIsArray(cfg.ADD_URI_SAFE_ATTR)) {
      addToSet(URI_SAFE_ATTRIBUTES, cfg.ADD_URI_SAFE_ATTR, transformCaseFunc);
    }
    if (objectHasOwnProperty(cfg, "FORBID_CONTENTS") && arrayIsArray(cfg.FORBID_CONTENTS)) {
      if (FORBID_CONTENTS === DEFAULT_FORBID_CONTENTS) {
        FORBID_CONTENTS = clone(FORBID_CONTENTS);
      }
      addToSet(FORBID_CONTENTS, cfg.FORBID_CONTENTS, transformCaseFunc);
    }
    if (objectHasOwnProperty(cfg, "ADD_FORBID_CONTENTS") && arrayIsArray(cfg.ADD_FORBID_CONTENTS)) {
      if (FORBID_CONTENTS === DEFAULT_FORBID_CONTENTS) {
        FORBID_CONTENTS = clone(FORBID_CONTENTS);
      }
      addToSet(FORBID_CONTENTS, cfg.ADD_FORBID_CONTENTS, transformCaseFunc);
    }
    if (KEEP_CONTENT) {
      ALLOWED_TAGS["#text"] = true;
    }
    if (WHOLE_DOCUMENT) {
      addToSet(ALLOWED_TAGS, ["html", "head", "body"]);
    }
    if (ALLOWED_TAGS.table) {
      addToSet(ALLOWED_TAGS, ["tbody"]);
      delete FORBID_TAGS.tbody;
    }
    if (cfg.TRUSTED_TYPES_POLICY) {
      if (typeof cfg.TRUSTED_TYPES_POLICY.createHTML !== "function") {
        throw typeErrorCreate('TRUSTED_TYPES_POLICY configuration option must provide a "createHTML" hook.');
      }
      if (typeof cfg.TRUSTED_TYPES_POLICY.createScriptURL !== "function") {
        throw typeErrorCreate('TRUSTED_TYPES_POLICY configuration option must provide a "createScriptURL" hook.');
      }
      trustedTypesPolicy = cfg.TRUSTED_TYPES_POLICY;
      emptyHTML = trustedTypesPolicy.createHTML("");
    } else {
      if (trustedTypesPolicy === undefined) {
        trustedTypesPolicy = _createTrustedTypesPolicy(trustedTypes, currentScript);
      }
      if (trustedTypesPolicy !== null && typeof emptyHTML === "string") {
        emptyHTML = trustedTypesPolicy.createHTML("");
      }
    }
    if (freeze) {
      freeze(cfg);
    }
    CONFIG = cfg;
  };
  const ALL_SVG_TAGS = addToSet({}, [...svg$1, ...svgFilters, ...svgDisallowed]);
  const ALL_MATHML_TAGS = addToSet({}, [...mathMl$1, ...mathMlDisallowed]);
  const _checkValidNamespace = function _checkValidNamespace2(element) {
    let parent = getParentNode(element);
    if (!parent || !parent.tagName) {
      parent = {
        namespaceURI: NAMESPACE,
        tagName: "template"
      };
    }
    const tagName = stringToLowerCase(element.tagName);
    const parentTagName = stringToLowerCase(parent.tagName);
    if (!ALLOWED_NAMESPACES[element.namespaceURI]) {
      return false;
    }
    if (element.namespaceURI === SVG_NAMESPACE) {
      if (parent.namespaceURI === HTML_NAMESPACE) {
        return tagName === "svg";
      }
      if (parent.namespaceURI === MATHML_NAMESPACE) {
        return tagName === "svg" && (parentTagName === "annotation-xml" || MATHML_TEXT_INTEGRATION_POINTS[parentTagName]);
      }
      return Boolean(ALL_SVG_TAGS[tagName]);
    }
    if (element.namespaceURI === MATHML_NAMESPACE) {
      if (parent.namespaceURI === HTML_NAMESPACE) {
        return tagName === "math";
      }
      if (parent.namespaceURI === SVG_NAMESPACE) {
        return tagName === "math" && HTML_INTEGRATION_POINTS[parentTagName];
      }
      return Boolean(ALL_MATHML_TAGS[tagName]);
    }
    if (element.namespaceURI === HTML_NAMESPACE) {
      if (parent.namespaceURI === SVG_NAMESPACE && !HTML_INTEGRATION_POINTS[parentTagName]) {
        return false;
      }
      if (parent.namespaceURI === MATHML_NAMESPACE && !MATHML_TEXT_INTEGRATION_POINTS[parentTagName]) {
        return false;
      }
      return !ALL_MATHML_TAGS[tagName] && (COMMON_SVG_AND_HTML_ELEMENTS[tagName] || !ALL_SVG_TAGS[tagName]);
    }
    if (PARSER_MEDIA_TYPE === "application/xhtml+xml" && ALLOWED_NAMESPACES[element.namespaceURI]) {
      return true;
    }
    return false;
  };
  const _forceRemove = function _forceRemove2(node) {
    arrayPush(DOMPurify.removed, {
      element: node
    });
    try {
      getParentNode(node).removeChild(node);
    } catch (_) {
      remove(node);
    }
  };
  const _removeAttribute = function _removeAttribute2(name, element) {
    try {
      arrayPush(DOMPurify.removed, {
        attribute: element.getAttributeNode(name),
        from: element
      });
    } catch (_) {
      arrayPush(DOMPurify.removed, {
        attribute: null,
        from: element
      });
    }
    element.removeAttribute(name);
    if (name === "is") {
      if (RETURN_DOM || RETURN_DOM_FRAGMENT) {
        try {
          _forceRemove(element);
        } catch (_) {}
      } else {
        try {
          element.setAttribute(name, "");
        } catch (_) {}
      }
    }
  };
  const _initDocument = function _initDocument2(dirty) {
    let doc = null;
    let leadingWhitespace = null;
    if (FORCE_BODY) {
      dirty = "<remove></remove>" + dirty;
    } else {
      const matches = stringMatch(dirty, /^[\r\n\t ]+/);
      leadingWhitespace = matches && matches[0];
    }
    if (PARSER_MEDIA_TYPE === "application/xhtml+xml" && NAMESPACE === HTML_NAMESPACE) {
      dirty = '<html xmlns="http://www.w3.org/1999/xhtml"><head></head><body>' + dirty + "</body></html>";
    }
    const dirtyPayload = trustedTypesPolicy ? trustedTypesPolicy.createHTML(dirty) : dirty;
    if (NAMESPACE === HTML_NAMESPACE) {
      try {
        doc = new DOMParser().parseFromString(dirtyPayload, PARSER_MEDIA_TYPE);
      } catch (_) {}
    }
    if (!doc || !doc.documentElement) {
      doc = implementation.createDocument(NAMESPACE, "template", null);
      try {
        doc.documentElement.innerHTML = IS_EMPTY_INPUT ? emptyHTML : dirtyPayload;
      } catch (_) {}
    }
    const body = doc.body || doc.documentElement;
    if (dirty && leadingWhitespace) {
      body.insertBefore(document2.createTextNode(leadingWhitespace), body.childNodes[0] || null);
    }
    if (NAMESPACE === HTML_NAMESPACE) {
      return getElementsByTagName.call(doc, WHOLE_DOCUMENT ? "html" : "body")[0];
    }
    return WHOLE_DOCUMENT ? doc.documentElement : body;
  };
  const _createNodeIterator = function _createNodeIterator2(root) {
    return createNodeIterator.call(root.ownerDocument || root, root, NodeFilter.SHOW_ELEMENT | NodeFilter.SHOW_COMMENT | NodeFilter.SHOW_TEXT | NodeFilter.SHOW_PROCESSING_INSTRUCTION | NodeFilter.SHOW_CDATA_SECTION, null);
  };
  const _isClobbered = function _isClobbered2(element) {
    return element instanceof HTMLFormElement && (typeof element.nodeName !== "string" || typeof element.textContent !== "string" || typeof element.removeChild !== "function" || !(element.attributes instanceof NamedNodeMap) || typeof element.removeAttribute !== "function" || typeof element.setAttribute !== "function" || typeof element.namespaceURI !== "string" || typeof element.insertBefore !== "function" || typeof element.hasChildNodes !== "function");
  };
  const _isNode = function _isNode2(value) {
    return typeof Node === "function" && value instanceof Node;
  };
  function _executeHooks(hooks2, currentNode, data) {
    arrayForEach(hooks2, (hook) => {
      hook.call(DOMPurify, currentNode, data, CONFIG);
    });
  }
  const _sanitizeElements = function _sanitizeElements2(currentNode) {
    let content = null;
    _executeHooks(hooks.beforeSanitizeElements, currentNode, null);
    if (_isClobbered(currentNode)) {
      _forceRemove(currentNode);
      return true;
    }
    const tagName = transformCaseFunc(currentNode.nodeName);
    _executeHooks(hooks.uponSanitizeElement, currentNode, {
      tagName,
      allowedTags: ALLOWED_TAGS
    });
    if (SAFE_FOR_XML && currentNode.hasChildNodes() && !_isNode(currentNode.firstElementChild) && regExpTest(/<[/\w!]/g, currentNode.innerHTML) && regExpTest(/<[/\w!]/g, currentNode.textContent)) {
      _forceRemove(currentNode);
      return true;
    }
    if (SAFE_FOR_XML && currentNode.namespaceURI === HTML_NAMESPACE && tagName === "style" && _isNode(currentNode.firstElementChild)) {
      _forceRemove(currentNode);
      return true;
    }
    if (currentNode.nodeType === NODE_TYPE.progressingInstruction) {
      _forceRemove(currentNode);
      return true;
    }
    if (SAFE_FOR_XML && currentNode.nodeType === NODE_TYPE.comment && regExpTest(/<[/\w]/g, currentNode.data)) {
      _forceRemove(currentNode);
      return true;
    }
    if (FORBID_TAGS[tagName] || !(EXTRA_ELEMENT_HANDLING.tagCheck instanceof Function && EXTRA_ELEMENT_HANDLING.tagCheck(tagName)) && !ALLOWED_TAGS[tagName]) {
      if (!FORBID_TAGS[tagName] && _isBasicCustomElement(tagName)) {
        if (CUSTOM_ELEMENT_HANDLING.tagNameCheck instanceof RegExp && regExpTest(CUSTOM_ELEMENT_HANDLING.tagNameCheck, tagName)) {
          return false;
        }
        if (CUSTOM_ELEMENT_HANDLING.tagNameCheck instanceof Function && CUSTOM_ELEMENT_HANDLING.tagNameCheck(tagName)) {
          return false;
        }
      }
      if (KEEP_CONTENT && !FORBID_CONTENTS[tagName]) {
        const parentNode = getParentNode(currentNode) || currentNode.parentNode;
        const childNodes = getChildNodes(currentNode) || currentNode.childNodes;
        if (childNodes && parentNode) {
          const childCount = childNodes.length;
          for (let i = childCount - 1;i >= 0; --i) {
            const childClone = cloneNode(childNodes[i], true);
            parentNode.insertBefore(childClone, getNextSibling(currentNode));
          }
        }
      }
      _forceRemove(currentNode);
      return true;
    }
    if (currentNode instanceof Element && !_checkValidNamespace(currentNode)) {
      _forceRemove(currentNode);
      return true;
    }
    if ((tagName === "noscript" || tagName === "noembed" || tagName === "noframes") && regExpTest(/<\/no(script|embed|frames)/i, currentNode.innerHTML)) {
      _forceRemove(currentNode);
      return true;
    }
    if (SAFE_FOR_TEMPLATES && currentNode.nodeType === NODE_TYPE.text) {
      content = currentNode.textContent;
      arrayForEach([MUSTACHE_EXPR2, ERB_EXPR2, TMPLIT_EXPR2], (expr) => {
        content = stringReplace(content, expr, " ");
      });
      if (currentNode.textContent !== content) {
        arrayPush(DOMPurify.removed, {
          element: currentNode.cloneNode()
        });
        currentNode.textContent = content;
      }
    }
    _executeHooks(hooks.afterSanitizeElements, currentNode, null);
    return false;
  };
  const _isValidAttribute = function _isValidAttribute2(lcTag, lcName, value) {
    if (FORBID_ATTR[lcName]) {
      return false;
    }
    if (SANITIZE_DOM && (lcName === "id" || lcName === "name") && ((value in document2) || (value in formElement))) {
      return false;
    }
    const nameIsPermitted = ALLOWED_ATTR[lcName] || EXTRA_ELEMENT_HANDLING.attributeCheck instanceof Function && EXTRA_ELEMENT_HANDLING.attributeCheck(lcName, lcTag);
    if (ALLOW_DATA_ATTR && !FORBID_ATTR[lcName] && regExpTest(DATA_ATTR2, lcName))
      ;
    else if (ALLOW_ARIA_ATTR && regExpTest(ARIA_ATTR2, lcName))
      ;
    else if (!nameIsPermitted || FORBID_ATTR[lcName]) {
      if (_isBasicCustomElement(lcTag) && (CUSTOM_ELEMENT_HANDLING.tagNameCheck instanceof RegExp && regExpTest(CUSTOM_ELEMENT_HANDLING.tagNameCheck, lcTag) || CUSTOM_ELEMENT_HANDLING.tagNameCheck instanceof Function && CUSTOM_ELEMENT_HANDLING.tagNameCheck(lcTag)) && (CUSTOM_ELEMENT_HANDLING.attributeNameCheck instanceof RegExp && regExpTest(CUSTOM_ELEMENT_HANDLING.attributeNameCheck, lcName) || CUSTOM_ELEMENT_HANDLING.attributeNameCheck instanceof Function && CUSTOM_ELEMENT_HANDLING.attributeNameCheck(lcName, lcTag)) || lcName === "is" && CUSTOM_ELEMENT_HANDLING.allowCustomizedBuiltInElements && (CUSTOM_ELEMENT_HANDLING.tagNameCheck instanceof RegExp && regExpTest(CUSTOM_ELEMENT_HANDLING.tagNameCheck, value) || CUSTOM_ELEMENT_HANDLING.tagNameCheck instanceof Function && CUSTOM_ELEMENT_HANDLING.tagNameCheck(value)))
        ;
      else {
        return false;
      }
    } else if (URI_SAFE_ATTRIBUTES[lcName])
      ;
    else if (regExpTest(IS_ALLOWED_URI$1, stringReplace(value, ATTR_WHITESPACE2, "")))
      ;
    else if ((lcName === "src" || lcName === "xlink:href" || lcName === "href") && lcTag !== "script" && stringIndexOf(value, "data:") === 0 && DATA_URI_TAGS[lcTag])
      ;
    else if (ALLOW_UNKNOWN_PROTOCOLS && !regExpTest(IS_SCRIPT_OR_DATA2, stringReplace(value, ATTR_WHITESPACE2, "")))
      ;
    else if (value) {
      return false;
    } else
      ;
    return true;
  };
  const RESERVED_CUSTOM_ELEMENT_NAMES = addToSet({}, ["annotation-xml", "color-profile", "font-face", "font-face-format", "font-face-name", "font-face-src", "font-face-uri", "missing-glyph"]);
  const _isBasicCustomElement = function _isBasicCustomElement2(tagName) {
    return !RESERVED_CUSTOM_ELEMENT_NAMES[stringToLowerCase(tagName)] && regExpTest(CUSTOM_ELEMENT2, tagName);
  };
  const _sanitizeAttributes = function _sanitizeAttributes2(currentNode) {
    _executeHooks(hooks.beforeSanitizeAttributes, currentNode, null);
    const {
      attributes
    } = currentNode;
    if (!attributes || _isClobbered(currentNode)) {
      return;
    }
    const hookEvent = {
      attrName: "",
      attrValue: "",
      keepAttr: true,
      allowedAttributes: ALLOWED_ATTR,
      forceKeepAttr: undefined
    };
    let l = attributes.length;
    while (l--) {
      const attr = attributes[l];
      const {
        name,
        namespaceURI,
        value: attrValue
      } = attr;
      const lcName = transformCaseFunc(name);
      const initValue = attrValue;
      let value = name === "value" ? initValue : stringTrim(initValue);
      hookEvent.attrName = lcName;
      hookEvent.attrValue = value;
      hookEvent.keepAttr = true;
      hookEvent.forceKeepAttr = undefined;
      _executeHooks(hooks.uponSanitizeAttribute, currentNode, hookEvent);
      value = hookEvent.attrValue;
      if (SANITIZE_NAMED_PROPS && (lcName === "id" || lcName === "name") && stringIndexOf(value, SANITIZE_NAMED_PROPS_PREFIX) !== 0) {
        _removeAttribute(name, currentNode);
        value = SANITIZE_NAMED_PROPS_PREFIX + value;
      }
      if (SAFE_FOR_XML && regExpTest(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, value)) {
        _removeAttribute(name, currentNode);
        continue;
      }
      if (lcName === "attributename" && stringMatch(value, "href")) {
        _removeAttribute(name, currentNode);
        continue;
      }
      if (hookEvent.forceKeepAttr) {
        continue;
      }
      if (!hookEvent.keepAttr) {
        _removeAttribute(name, currentNode);
        continue;
      }
      if (!ALLOW_SELF_CLOSE_IN_ATTR && regExpTest(/\/>/i, value)) {
        _removeAttribute(name, currentNode);
        continue;
      }
      if (SAFE_FOR_TEMPLATES) {
        arrayForEach([MUSTACHE_EXPR2, ERB_EXPR2, TMPLIT_EXPR2], (expr) => {
          value = stringReplace(value, expr, " ");
        });
      }
      const lcTag = transformCaseFunc(currentNode.nodeName);
      if (!_isValidAttribute(lcTag, lcName, value)) {
        _removeAttribute(name, currentNode);
        continue;
      }
      if (trustedTypesPolicy && typeof trustedTypes === "object" && typeof trustedTypes.getAttributeType === "function") {
        if (namespaceURI)
          ;
        else {
          switch (trustedTypes.getAttributeType(lcTag, lcName)) {
            case "TrustedHTML": {
              value = trustedTypesPolicy.createHTML(value);
              break;
            }
            case "TrustedScriptURL": {
              value = trustedTypesPolicy.createScriptURL(value);
              break;
            }
          }
        }
      }
      if (value !== initValue) {
        try {
          if (namespaceURI) {
            currentNode.setAttributeNS(namespaceURI, name, value);
          } else {
            currentNode.setAttribute(name, value);
          }
          if (_isClobbered(currentNode)) {
            _forceRemove(currentNode);
          } else {
            arrayPop(DOMPurify.removed);
          }
        } catch (_) {
          _removeAttribute(name, currentNode);
        }
      }
    }
    _executeHooks(hooks.afterSanitizeAttributes, currentNode, null);
  };
  const _sanitizeShadowDOM2 = function _sanitizeShadowDOM(fragment) {
    let shadowNode = null;
    const shadowIterator = _createNodeIterator(fragment);
    _executeHooks(hooks.beforeSanitizeShadowDOM, fragment, null);
    while (shadowNode = shadowIterator.nextNode()) {
      _executeHooks(hooks.uponSanitizeShadowNode, shadowNode, null);
      _sanitizeElements(shadowNode);
      _sanitizeAttributes(shadowNode);
      if (shadowNode.content instanceof DocumentFragment) {
        _sanitizeShadowDOM2(shadowNode.content);
      }
    }
    _executeHooks(hooks.afterSanitizeShadowDOM, fragment, null);
  };
  DOMPurify.sanitize = function(dirty) {
    let cfg = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {};
    let body = null;
    let importedNode = null;
    let currentNode = null;
    let returnNode = null;
    IS_EMPTY_INPUT = !dirty;
    if (IS_EMPTY_INPUT) {
      dirty = "<!-->";
    }
    if (typeof dirty !== "string" && !_isNode(dirty)) {
      dirty = stringifyValue(dirty);
      if (typeof dirty !== "string") {
        throw typeErrorCreate("dirty is not a string, aborting");
      }
    }
    if (!DOMPurify.isSupported) {
      return dirty;
    }
    if (!SET_CONFIG) {
      _parseConfig(cfg);
    }
    DOMPurify.removed = [];
    if (typeof dirty === "string") {
      IN_PLACE = false;
    }
    if (IN_PLACE) {
      const nn = dirty.nodeName;
      if (typeof nn === "string") {
        const tagName = transformCaseFunc(nn);
        if (!ALLOWED_TAGS[tagName] || FORBID_TAGS[tagName]) {
          throw typeErrorCreate("root node is forbidden and cannot be sanitized in-place");
        }
      }
    } else if (dirty instanceof Node) {
      body = _initDocument("<!---->");
      importedNode = body.ownerDocument.importNode(dirty, true);
      if (importedNode.nodeType === NODE_TYPE.element && importedNode.nodeName === "BODY") {
        body = importedNode;
      } else if (importedNode.nodeName === "HTML") {
        body = importedNode;
      } else {
        body.appendChild(importedNode);
      }
    } else {
      if (!RETURN_DOM && !SAFE_FOR_TEMPLATES && !WHOLE_DOCUMENT && dirty.indexOf("<") === -1) {
        return trustedTypesPolicy && RETURN_TRUSTED_TYPE ? trustedTypesPolicy.createHTML(dirty) : dirty;
      }
      body = _initDocument(dirty);
      if (!body) {
        return RETURN_DOM ? null : RETURN_TRUSTED_TYPE ? emptyHTML : "";
      }
    }
    if (body && FORCE_BODY) {
      _forceRemove(body.firstChild);
    }
    const nodeIterator = _createNodeIterator(IN_PLACE ? dirty : body);
    while (currentNode = nodeIterator.nextNode()) {
      _sanitizeElements(currentNode);
      _sanitizeAttributes(currentNode);
      if (currentNode.content instanceof DocumentFragment) {
        _sanitizeShadowDOM2(currentNode.content);
      }
    }
    if (IN_PLACE) {
      return dirty;
    }
    if (RETURN_DOM) {
      if (SAFE_FOR_TEMPLATES) {
        body.normalize();
        let html2 = body.innerHTML;
        arrayForEach([MUSTACHE_EXPR2, ERB_EXPR2, TMPLIT_EXPR2], (expr) => {
          html2 = stringReplace(html2, expr, " ");
        });
        body.innerHTML = html2;
      }
      if (RETURN_DOM_FRAGMENT) {
        returnNode = createDocumentFragment.call(body.ownerDocument);
        while (body.firstChild) {
          returnNode.appendChild(body.firstChild);
        }
      } else {
        returnNode = body;
      }
      if (ALLOWED_ATTR.shadowroot || ALLOWED_ATTR.shadowrootmode) {
        returnNode = importNode.call(originalDocument, returnNode, true);
      }
      return returnNode;
    }
    let serializedHTML = WHOLE_DOCUMENT ? body.outerHTML : body.innerHTML;
    if (WHOLE_DOCUMENT && ALLOWED_TAGS["!doctype"] && body.ownerDocument && body.ownerDocument.doctype && body.ownerDocument.doctype.name && regExpTest(DOCTYPE_NAME, body.ownerDocument.doctype.name)) {
      serializedHTML = "<!DOCTYPE " + body.ownerDocument.doctype.name + `>
` + serializedHTML;
    }
    if (SAFE_FOR_TEMPLATES) {
      arrayForEach([MUSTACHE_EXPR2, ERB_EXPR2, TMPLIT_EXPR2], (expr) => {
        serializedHTML = stringReplace(serializedHTML, expr, " ");
      });
    }
    return trustedTypesPolicy && RETURN_TRUSTED_TYPE ? trustedTypesPolicy.createHTML(serializedHTML) : serializedHTML;
  };
  DOMPurify.setConfig = function() {
    let cfg = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};
    _parseConfig(cfg);
    SET_CONFIG = true;
  };
  DOMPurify.clearConfig = function() {
    CONFIG = null;
    SET_CONFIG = false;
  };
  DOMPurify.isValidAttribute = function(tag, attr, value) {
    if (!CONFIG) {
      _parseConfig({});
    }
    const lcTag = transformCaseFunc(tag);
    const lcName = transformCaseFunc(attr);
    return _isValidAttribute(lcTag, lcName, value);
  };
  DOMPurify.addHook = function(entryPoint, hookFunction) {
    if (typeof hookFunction !== "function") {
      return;
    }
    arrayPush(hooks[entryPoint], hookFunction);
  };
  DOMPurify.removeHook = function(entryPoint, hookFunction) {
    if (hookFunction !== undefined) {
      const index = arrayLastIndexOf(hooks[entryPoint], hookFunction);
      return index === -1 ? undefined : arraySplice(hooks[entryPoint], index, 1)[0];
    }
    return arrayPop(hooks[entryPoint]);
  };
  DOMPurify.removeHooks = function(entryPoint) {
    hooks[entryPoint] = [];
  };
  DOMPurify.removeAllHooks = function() {
    hooks = _createHooksMap();
  };
  return DOMPurify;
}
var purify = createDOMPurify();

// src/client/shared/utils.ts
function showLoader(show) {
  const loader = document.getElementById("loader");
  if (loader) {
    loader.classList.toggle("hidden", !show);
  }
}
function formatDate(dateStr) {
  if (!dateStr)
    return "";
  const d = new Date(dateStr);
  if (Number.isNaN(d.getTime()))
    return "";
  const pad = (n) => n < 10 ? `0${n}` : n;
  const day = pad(d.getDate());
  const month = pad(d.getMonth() + 1);
  const year = d.getFullYear();
  return `${day}/${month}/${year}`;
}
function playBeep(type = "success") {
  try {
    const AudioContext = window.AudioContext || window.webkitAudioContext;
    if (!AudioContext)
      return;
    const ctx = new AudioContext;
    const playTone = (freq, duration, start) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.type = "sine";
      osc.frequency.setValueAtTime(freq, ctx.currentTime + start);
      gain.gain.setValueAtTime(0.1, ctx.currentTime + start);
      osc.start(ctx.currentTime + start);
      osc.stop(ctx.currentTime + start + duration);
    };
    if (type === "success")
      playTone(880, 0.15, 0);
    else if (type === "warning") {
      playTone(660, 0.1, 0);
      playTone(660, 0.1, 0.15);
    } else if (type === "error")
      playTone(440, 0.4, 0);
  } catch (_e) {}
}
function escHtml(unsafe) {
  if (unsafe == null)
    return "";
  return String(unsafe).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#039;");
}
function sanitizeHtml(dirty) {
  return purify.sanitize(dirty);
}

// src/client/app/state.ts
var state = {
  screen: localStorage.getItem("simp_screen") || "login",
  user: JSON.parse(localStorage.getItem("simp_user") || "null"),
  token: null,
  atividade: JSON.parse(localStorage.getItem("simp_atividade") || "null"),
  scannedProducts: JSON.parse(localStorage.getItem("simp_scanned") || "[]"),
  expectedProducts: JSON.parse(localStorage.getItem("simp_expected") || "[]"),
  lastScanned: null,
  allowKeyboard: false
};
function saveState() {
  localStorage.setItem("simp_screen", state.screen);
  if (state.user)
    localStorage.setItem("simp_user", JSON.stringify(state.user));
  else
    localStorage.removeItem("simp_user");
  if (state.atividade)
    localStorage.setItem("simp_atividade", JSON.stringify(state.atividade));
  else
    localStorage.removeItem("simp_atividade");
  localStorage.setItem("simp_scanned", JSON.stringify(state.scannedProducts));
  localStorage.setItem("simp_expected", JSON.stringify(state.expectedProducts));
}
function resetActivityState() {
  state.atividade = null;
  state.scannedProducts = [];
  state.expectedProducts = [];
  saveState();
}

// src/client/app/ui.ts
function focusScannerInput(input) {
  if (!input)
    return;
  if (state.allowKeyboard) {
    input.readOnly = false;
    input.focus();
    return;
  }
  input.readOnly = true;
  input.focus();
  setTimeout(() => {
    input.readOnly = false;
  }, 50);
}
function refocusInput() {
  if (state.screen === "scanning" || state.screen === "consulta") {
    const inputId = state.screen === "scanning" ? "scan-input" : "consulta-input";
    const input = document.getElementById(inputId);
    const reauthModal = document.getElementById("modal-reauth");
    const isReauthVisible = reauthModal && !reauthModal.classList.contains("hidden");
    if (state.allowKeyboard && document.activeElement === input)
      return;
    if (input && document.activeElement !== input && !isReauthVisible) {
      const activeTag = document.activeElement ? document.activeElement.tagName.toLowerCase() : "";
      if (!["select", "textarea", "input"].includes(activeTag)) {
        focusScannerInput(input);
      }
    }
  }
}
function syncKeyboardUI() {
  const screens = [
    { btnId: "btn-toggle-keyboard", inputId: "scan-input" },
    { btnId: "btn-consulta-toggle-keyboard", inputId: "consulta-input" }
  ];
  for (const s of screens) {
    const btn = document.getElementById(s.btnId);
    const input = document.getElementById(s.inputId);
    if (btn)
      btn.classList.toggle("active", state.allowKeyboard);
    if (input) {
      input.inputMode = state.allowKeyboard ? "tel" : "none";
      if (state.allowKeyboard && state.screen === (s.inputId === "scan-input" ? "scanning" : "consulta")) {
        input.focus();
      } else if (!state.allowKeyboard && state.screen === (s.inputId === "scan-input" ? "scanning" : "consulta")) {
        focusScannerInput(input);
      }
    }
  }
}
function syncReposicaoUI() {}
function renderHistory() {
  const histEl = document.getElementById("scan-history");
  if (!histEl)
    return;
  const unique = [];
  const reversed = [...state.scannedProducts].reverse();
  for (const p of reversed) {
    if (!unique.find((x) => x.seqproduto === p.seqproduto))
      unique.push(p);
  }
  const displayList = unique;
  histEl.innerHTML = sanitizeHtml(displayList.map((p) => `
        <div class="history-item" data-seqproduto="${p.seqproduto}">
            <span class="truncate" style="flex:1; margin-right: 8px;">${p.desccompleta || p.ean}</span>
            <div style="display: flex; gap: 4px; align-items: center;">
                ${p.reposicao ? '<span title="Reposição">\uD83D\uDCE6</span>' : ""}
                <span class="${p.status === "OK" ? "status-ok" : "status-warn"}">${p.status === "OK" ? "✔️" : "⚠️"}</span>
            </div>
        </div>
    `).join(""));
}
function showProductDetailModal(seqproduto) {
  const product = state.scannedProducts.find((p) => p.seqproduto === seqproduto);
  if (!product)
    return;
  const modal = document.getElementById("modal-product-detail");
  if (!modal)
    return;
  const descEl = document.getElementById("product-detail-desc");
  const eanEl = document.getElementById("product-detail-ean");
  const seqEl = document.getElementById("product-detail-seq");
  const statusEl = document.getElementById("product-detail-status");
  const localEl = document.getElementById("product-detail-local");
  if (descEl)
    descEl.innerText = product.desccompleta || "Sem descrição";
  if (eanEl)
    eanEl.innerText = product.ean || "N/A";
  if (seqEl)
    seqEl.innerText = String(product.seqproduto);
  if (statusEl)
    statusEl.innerText = product.status || "N/A";
  if (localEl)
    localEl.innerText = product.rua != null && product.predio != null ? `${product.rua}/${product.predio}` : "N/A";
  const toggleBtn = document.getElementById("btn-modal-toggle-reposicao");
  if (toggleBtn) {
    if (product.reposicao) {
      toggleBtn.innerHTML = sanitizeHtml('Remover Reposição <span style="filter: grayscale(0);">\uD83D\uDCE6</span>');
      toggleBtn.style.backgroundColor = "#fef3c7";
      toggleBtn.style.color = "#92400e";
      toggleBtn.style.borderColor = "#f59e0b";
    } else {
      toggleBtn.innerHTML = sanitizeHtml('Marcar Reposição <span style="filter: grayscale(1); opacity: 0.5;">\uD83D\uDCE6</span>');
      toggleBtn.style.backgroundColor = "#f8fafc";
      toggleBtn.style.color = "#64748b";
      toggleBtn.style.borderColor = "#e2e8f0";
    }
    toggleBtn.onclick = () => {
      product.reposicao = !product.reposicao;
      saveState();
      renderHistory();
      showProductDetailModal(seqproduto);
    };
  }
  const removeBtn = document.getElementById("btn-modal-remove");
  if (removeBtn) {
    const canRemove = state.atividade !== null && state.user !== null;
    removeBtn.style.display = canRemove ? "block" : "none";
    if (canRemove) {
      removeBtn.onclick = () => {
        if (!confirm("Remover este produto da leitura?")) {
          return;
        }
        state.scannedProducts = state.scannedProducts.filter((p) => p.seqproduto !== seqproduto);
        saveState();
        renderHistory();
        closeProductDetailModal();
      };
    }
  }
  modal.classList.remove("hidden");
}
function closeProductDetailModal() {
  document.getElementById("modal-product-detail")?.classList.add("hidden");
}

// src/client/app/navigation.ts
var _loadEmpresasCb = null;
function setLoadEmpresasCb(cb) {
  _loadEmpresasCb = cb;
}
function showScreen(screenId) {
  if (screenId === "consulta") {
    state.previousScreen = state.screen;
  }
  state.screen = screenId;
  saveState();
  const screens = document.querySelectorAll(".screen");
  for (let i = 0;i < screens.length; i++) {
    screens[i].classList.add("hidden");
  }
  const targetScreen = document.getElementById(`screen-${screenId}`);
  if (targetScreen)
    targetScreen.classList.remove("hidden");
  syncKeyboardUI();
  syncReposicaoUI();
  const isAuth = screenId !== "login";
  const hideGlobalHeader = [
    "scanning",
    "divergence",
    "predio-switch",
    "consulta"
  ].includes(screenId);
  const headerActions = document.getElementById("header-actions");
  if (headerActions)
    headerActions.classList.toggle("hidden", !isAuth);
  const header = document.querySelector(".header");
  if (header)
    header.classList.toggle("hidden", hideGlobalHeader);
  if (isAuth) {
    const userName = state.user?.nome || state.user?.username || "Usuário";
    const displayUser = userName.length > 12 ? `${userName.substring(0, 10)}..` : userName;
    const userEl = document.getElementById("header-user");
    if (userEl)
      userEl.innerText = displayUser;
  }
  if (screenId === "start") {
    const infoEl = document.getElementById("last-activity-info");
    if (infoEl)
      infoEl.innerText = "";
    _loadEmpresasCb?.();
  }
  if (screenId === "scanning" && state.atividade) {
    const currentPredio = state.atividade.currentPredio || state.atividade.predio;
    const userName = state.user?.nome || state.user?.username || "Usuário";
    const displayUser = userName.length > 12 ? `${userName.substring(0, 10)}..` : userName;
    const scanAddress = document.getElementById("scan-address");
    if (scanAddress)
      scanAddress.innerText = `${state.atividade.rua} | ${currentPredio} • ${displayUser}`;
    const scanFeedback = document.getElementById("scan-feedback");
    if (scanFeedback)
      scanFeedback.innerHTML = "";
    renderHistory();
    setTimeout(refocusInput, 100);
  }
}
async function confirmExit(callback) {
  if (state.scannedProducts.length > 0 && ["scanning", "divergence", "predio-switch"].includes(state.screen)) {
    if (!confirm("Você tem produtos lidos. Deseja realmente sair SEM salvar a atividade? Os dados lidos serão perdidos.")) {
      return;
    }
  }
  callback();
}

// src/client/app/auth.ts
function logout() {
  confirmExit(() => {
    fetch("/api/auth/logout", { method: "POST", credentials: "include" }).catch(() => {});
    state.user = null;
    state.token = null;
    saveState();
    showScreen("login");
  });
}
function showReauthModal(show) {
  const modal = document.getElementById("modal-reauth");
  if (!modal)
    return;
  modal.classList.toggle("hidden", !show);
  if (show) {
    const passInput = document.getElementById("reauth-password");
    if (passInput) {
      passInput.value = "";
      passInput.focus();
    }
    const errEl = document.getElementById("reauth-error");
    if (errEl)
      errEl.classList.add("hidden");
  }
}

// src/client/app/api.ts
async function loadEmpresas() {
  const { ok, data } = await apiCall("/api/empresas", {}, () => showReauthModal(true));
  if (ok && Array.isArray(data)) {
    const select = document.getElementById("start-empresa");
    if (select) {
      select.innerHTML = sanitizeHtml(data.map((e) => `<option value="${String(e.NROEMPRESA)}">${String(e.NROEMPRESA)} - ${String(e.NOMEREDUZIDO)}</option>`).join(""));
      if (data.length > 0) {
        loadLocais(data[0].NROEMPRESA);
      }
    }
  }
}
async function loadLocais(empresaId) {
  const { ok, data } = await apiCall(`/api/locais?empresa=${empresaId}`, {}, () => showReauthModal(true));
  const select = document.getElementById("start-local");
  if (select) {
    if (ok && Array.isArray(data) && data.length > 0) {
      select.innerHTML = sanitizeHtml(data.map((e) => `<option value="${String(e.SEQLOCAL)}">${String(e.LOCAL)}</option>`).join(""));
    } else {
      select.innerHTML = sanitizeHtml('<option value="">Nenhum local ativo</option>');
    }
  }
}
async function fetchLastActivityInfo() {
  const empresa = document.getElementById("start-empresa")?.value;
  const seqlocal = document.getElementById("start-local")?.value;
  const rua = document.getElementById("start-rua")?.value;
  const predio = document.getElementById("start-predio")?.value;
  const infoEl = document.getElementById("last-activity-info");
  if (!infoEl)
    return;
  if (!state.token)
    return;
  if (!empresa || !seqlocal || !rua || !predio) {
    infoEl.innerText = "";
    return;
  }
  infoEl.innerText = "Buscando histórico...";
  const res = await apiCall(`/api/atividades/last-info?empresa=${empresa}&seqlocal=${seqlocal}&rua=${rua}&predio=${predio}`, {}, () => showReauthModal(true));
  if (res.ok && res.data?.dataFim) {
    infoEl.innerHTML = sanitizeHtml(`Data última atividade: <strong style="color: #4f46e5;">${formatDate(res.data.dataFim)}</strong>`);
  } else if (res.ok && !res.data) {
    infoEl.innerText = "Nenhuma atividade encontrada";
  } else {
    infoEl.innerText = "";
  }
}

// src/client/app/scan.ts
async function startActivity() {
  const empSelect = document.getElementById("start-empresa");
  const empresa = empSelect.value;
  const empresaNome = empSelect.options[empSelect.selectedIndex].text.split(" - ")[1] || empSelect.options[empSelect.selectedIndex].text;
  const seqlocal = document.getElementById("start-local").value;
  const rua = document.getElementById("start-rua").value;
  const predio = document.getElementById("start-predio").value;
  const errEl = document.getElementById("start-error");
  if (!seqlocal) {
    if (errEl) {
      errEl.innerText = "Selecione um local válido";
      errEl.classList.remove("hidden");
    }
    return;
  }
  showLoader(true);
  const { ok, data } = await apiCall(`/api/produtos/local?empresa=${empresa}&seqlocal=${seqlocal}&rua=${rua}&predio=${predio}`, {}, () => showReauthModal(true));
  showLoader(false);
  if (ok && Array.isArray(data) && data.length > 0) {
    state.expectedProducts = data;
    state.atividade = {
      id: 0,
      empresa,
      empresaNome,
      seqlocal,
      rua,
      predio,
      predios: [predio],
      currentPredio: predio,
      status: "aberta",
      dataInicio: new Date().toISOString()
    };
    state.scannedProducts = [];
    saveState();
    showScreen("scanning");
  } else {
    if (errEl) {
      errEl.innerText = "Endereço não encontrado ou sem produtos";
      errEl.classList.remove("hidden");
    }
  }
}
async function finalizeActivity() {
  if (!state.atividade)
    return;
  if (!confirm("Tem certeza que deseja finalizar a atividade?")) {
    return;
  }
  showLoader(true);
  const predios = state.atividade.predios || [state.atividade.predio];
  const payload = {
    empresa: state.atividade.empresa,
    seqlocal: state.atividade.seqlocal,
    rua: state.atividade.rua,
    predio: predios,
    readProducts: state.scannedProducts,
    expectedProducts: state.expectedProducts
  };
  const result = await apiCall("/api/atividades/finalizar", {
    method: "POST",
    body: JSON.stringify(payload)
  }, () => showReauthModal(true));
  showLoader(false);
  if (result.ok) {
    resetActivityState();
    const rp = result.data || {};
    const divergences = rp.divergences || [];
    const ruptures = rp.ruptures || [];
    const replenishments = rp.replenishments || [];
    const reportIdEl = document.getElementById("report-id");
    const countDivEl = document.getElementById("count-div");
    const countRupEl = document.getElementById("count-rup");
    const countRepEl = document.getElementById("count-rep");
    if (reportIdEl)
      reportIdEl.innerText = rp.atividadeId || "--";
    if (countDivEl)
      countDivEl.innerText = divergences.length.toString();
    if (countRupEl)
      countRupEl.innerText = ruptures.length.toString();
    if (countRepEl)
      countRepEl.innerText = replenishments.length.toString();
    const divEl = document.getElementById("report-divergences");
    const rupEl = document.getElementById("report-ruptures");
    const repEl = document.getElementById("report-replenishments");
    const itemHtml = (p) => `<div style="padding: 0.5rem 0; border-bottom: 1px solid rgba(0,0,0,0.05);">
            <strong style="color: #334155;">SEQ: ${p.seqproduto || p.ean || "-"}</strong>
            <p style="color: #64748b; margin-top: 0.25rem;">${p.desccompleta || "-"}</p>
        </div>`;
    if (divEl)
      divEl.innerHTML = sanitizeHtml(divergences.map(itemHtml).join("") || '<p style="color: #94a3b8; font-style: italic;">Nenhuma divergência</p>');
    if (rupEl)
      rupEl.innerHTML = sanitizeHtml(ruptures.map(itemHtml).join("") || '<p style="color: #94a3b8; font-style: italic;">Nenhuma ruptura</p>');
    if (repEl)
      repEl.innerHTML = sanitizeHtml(replenishments.map(itemHtml).join("") || '<p style="color: #94a3b8; font-style: italic;">Nenhuma reposição</p>');
    showScreen("report");
  } else {
    if (result.status === 401) {
      return;
    }
    alert("Erro ao finalizar: " + (result.data.error || "Erro de conexão") + `

Seus dados estão salvos localmente. Tente novamente quando recuperar o sinal.`);
  }
}

// src/client/app/session.ts
function resolveAtividadesEntryScreen(hasUser, screen, hasAtividade) {
  if (!hasUser)
    return "login";
  if (hasAtividade)
    return "scanning";
  if (screen === "login")
    return "start";
  return screen;
}

// src/client/app/index.ts
setLoadEmpresasCb(loadEmpresas);
window.showProductDetailModal = showProductDetailModal;
window.closeProductDetailModal = closeProductDetailModal;
document.getElementById("scan-history")?.addEventListener("click", (e) => {
  const target = e.target.closest(".history-item");
  if (target) {
    const seq = target.getAttribute("data-seqproduto");
    if (seq)
      showProductDetailModal(Number(seq));
  }
});
document.addEventListener("DOMContentLoaded", async () => {
  const sessionRes = await apiCall("/api/auth/me", {}, () => {});
  if (sessionRes.ok && sessionRes.data?.user) {
    state.token = "cookie";
    state.user = sessionRes.data.user;
    saveState();
  }
  showScreen(resolveAtividadesEntryScreen(Boolean(state.user), state.screen, Boolean(state.atividade)));
  const btnLogout = document.getElementById("btn-logout");
  if (btnLogout)
    btnLogout.addEventListener("click", logout);
  const logouts = document.querySelectorAll(".js-btn-logout");
  for (let i = 0;i < logouts.length; i++) {
    logouts[i].addEventListener("click", logout);
  }
  function preventKeyboard(e) {
    if (!state.allowKeyboard) {
      const target = e.currentTarget;
      if (target && (state.screen === "scanning" || state.screen === "consulta")) {
        e.preventDefault();
        focusScannerInput(target);
      }
    }
  }
  const scanInput = document.getElementById("scan-input");
  if (scanInput) {
    scanInput.addEventListener("blur", () => {
      setTimeout(refocusInput, 300);
    });
    scanInput.addEventListener("mousedown", preventKeyboard);
    scanInput.addEventListener("touchstart", preventKeyboard);
  }
  const consultaInput = document.getElementById("consulta-input");
  if (consultaInput) {
    consultaInput.addEventListener("blur", () => {
      setTimeout(refocusInput, 300);
    });
    consultaInput.addEventListener("mousedown", preventKeyboard);
    consultaInput.addEventListener("touchstart", preventKeyboard);
  }
  const handleGlobalFocusLock = (e) => {
    if (!state.allowKeyboard && (state.screen === "scanning" || state.screen === "consulta")) {
      const target = e.target;
      const isInteractive = ["BUTTON", "INPUT", "SELECT", "TEXTAREA", "A", "SPAN", "I"].includes(target.tagName) || target.closest("button") || target.closest(".history-item") || target.closest(".modal-content") || target.closest(".scan-history");
      if (!isInteractive && e.type !== "touchmove") {
        e.preventDefault();
      }
    }
  };
  document.addEventListener("touchstart", handleGlobalFocusLock, {
    passive: true
  });
  document.addEventListener("mousedown", handleGlobalFocusLock);
  const toggleKeyboard = () => {
    state.allowKeyboard = !state.allowKeyboard;
    syncKeyboardUI();
    saveState();
  };
  document.getElementById("btn-toggle-keyboard")?.addEventListener("click", (e) => {
    toggleKeyboard();
    if (e.currentTarget instanceof HTMLElement)
      e.currentTarget.blur();
    setTimeout(refocusInput, 100);
  });
  document.getElementById("btn-consulta-toggle-keyboard")?.addEventListener("click", (e) => {
    toggleKeyboard();
    if (e.currentTarget instanceof HTMLElement)
      e.currentTarget.blur();
    setTimeout(refocusInput, 100);
  });
  const toggles = document.querySelectorAll(".password-toggle");
  for (let i = 0;i < toggles.length; i++) {
    toggles[i].addEventListener("click", (e) => {
      const input = e.target.previousElementSibling;
      if (input && input.tagName === "INPUT") {
        if (input.type === "password") {
          input.type = "text";
          e.target.innerText = "\uD83D\uDE48";
        } else {
          input.type = "password";
          e.target.innerText = "\uD83D\uDC41️";
        }
      }
    });
  }
  document.getElementById("btn-back-to-start")?.addEventListener("click", () => {
    confirmExit(() => showScreen("start"));
  });
  document.getElementById("start-empresa")?.addEventListener("change", (e) => {
    loadLocais(e.target.value);
    fetchLastActivityInfo();
  });
  document.getElementById("start-local")?.addEventListener("change", fetchLastActivityInfo);
  document.getElementById("start-rua")?.addEventListener("blur", fetchLastActivityInfo);
  document.getElementById("start-predio")?.addEventListener("blur", fetchLastActivityInfo);
  document.getElementById("form-login")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    showLoader(true);
    const errEl = document.getElementById("login-error");
    if (errEl)
      errEl.classList.add("hidden");
    const { ok, data } = await apiCall("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({
        username: document.getElementById("login-username")?.value,
        password: document.getElementById("login-password")?.value
      })
    });
    showLoader(false);
    if (ok) {
      state.user = data.user;
      state.token = data.token;
      saveState();
      if (state.atividade) {
        showScreen("scanning");
      } else {
        showScreen("start");
      }
    } else {
      if (errEl) {
        errEl.innerText = data.error || "Erro ao logar";
        errEl.classList.remove("hidden");
      }
    }
  });
  document.getElementById("form-reauth")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    showLoader(true);
    const errEl = document.getElementById("reauth-error");
    if (errEl)
      errEl.classList.add("hidden");
    const { ok, data } = await apiCall("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({
        username: state.user?.username,
        password: document.getElementById("reauth-password")?.value
      })
    });
    showLoader(false);
    if (ok) {
      state.token = data.token;
      saveState();
      showReauthModal(false);
      alert("Sessão revalidada! Você pode continuar.");
    } else {
      if (errEl) {
        errEl.innerText = data.error || "Senha incorreta";
        errEl.classList.remove("hidden");
      }
    }
  });
  document.getElementById("btn-reauth-cancel")?.addEventListener("click", () => {
    if (confirm("Ao sair agora, todos os produtos lidos nesta atividade serão PERDIDOS. Tem certeza?")) {
      showReauthModal(false);
      state.user = null;
      state.token = null;
      saveState();
      showScreen("login");
    }
  });
  document.getElementById("form-start")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (state.atividade && state.scannedProducts.length > 0) {
      if (!confirm(`Você já possui uma atividade em andamento com produtos lidos.

Deseja DESCARTAR os dados e iniciar uma nova?

• Clique OK para descartar e começar nova.
• Clique CANCELAR para voltar à atividade em andamento.`)) {
        showScreen("scanning");
        return;
      }
      resetActivityState();
    }
    await startActivity();
  });
  document.getElementById("form-scan")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const input = document.getElementById("scan-input");
    const code = input.value.trim();
    if (!code)
      return;
    if (!state.atividade)
      return;
    showLoader(true);
    const { ok, data } = await apiCall(`/api/produtos/ean/${code}?empresa=${state.atividade.empresa}&seqlocal=${state.atividade.seqlocal}`, {}, () => showReauthModal(true));
    showLoader(false);
    const feedback = document.getElementById("scan-feedback");
    if (!feedback)
      return;
    if (!ok) {
      playBeep("error");
      feedback.innerHTML = sanitizeHtml(`<div style="color: #ef4444; font-weight: bold;">❌ Produto não encontrado</div>`);
      input.select();
      return;
    }
    input.value = "";
    focusScannerInput(input);
    const currentPredio = state.atividade.currentPredio || state.atividade.predio;
    const isNullAddress = data.rua == null || data.predio == null;
    const sameRua = data.rua === state.atividade.rua;
    const samePredio = String(data.predio) === String(currentPredio);
    const alreadyScanned = state.scannedProducts.some((p) => p.seqproduto === data.seqproduto);
    if (alreadyScanned) {
      playBeep("warning");
      feedback.innerHTML = sanitizeHtml(`<div style="color: #f59e0b; font-weight: bold;">⚠️ Produto já lido nesta atividade!</div>`);
      return;
    }
    let status = "OK";
    if (isNullAddress || !sameRua || !samePredio) {
      status = "DIVERGENTE";
    }
    state.scannedProducts.push({
      seqproduto: data.seqproduto,
      ean: code,
      rua: data.rua,
      predio: String(currentPredio),
      desccompleta: data.desccompleta,
      status,
      reposicao: false
    });
    saveState();
    state.lastScanned = data;
    if (isNullAddress || !sameRua) {
      playBeep("warning");
      const reason = isNullAddress ? "S/ Endereço" : "Rua Divergente";
      feedback.innerHTML = sanitizeHtml(`<div style="color: #f59e0b; font-weight: bold;">⚠️ ${reason}: ${data.desccompleta}</div>`);
      renderHistory();
    } else if (!samePredio) {
      playBeep("warning");
      const displayPredio = data.predio != null ? data.predio : "N/A";
      const predioSwitchDesc = document.getElementById("predio-switch-desc");
      if (predioSwitchDesc)
        predioSwitchDesc.innerText = `${data.desccompleta} pertence ao Prédio ${displayPredio} (mesma rua).`;
      const predioSwitchNew = document.getElementById("predio-switch-new");
      if (predioSwitchNew)
        predioSwitchNew.innerText = displayPredio.toString();
      const predioSwitchCurrent = document.getElementById("predio-switch-current");
      if (predioSwitchCurrent)
        predioSwitchCurrent.innerText = currentPredio.toString();
      renderHistory();
      showScreen("predio-switch");
    } else {
      playBeep("success");
      feedback.innerHTML = sanitizeHtml(`<div style="color: #10b981; font-weight: bold;">✅ Lido: ${data.desccompleta}</div>`);
      renderHistory();
    }
  });
  document.getElementById("btn-finalize")?.addEventListener("click", finalizeActivity);
  document.getElementById("btn-div-continue")?.addEventListener("click", () => {
    showScreen("scanning");
    const feedback = document.getElementById("scan-feedback");
    if (feedback && state.lastScanned) {
      const lastScannedFromList = [...state.scannedProducts].reverse().find((p) => p.seqproduto === state.lastScanned?.seqproduto);
      const isRep = lastScannedFromList?.reposicao;
      feedback.innerHTML = sanitizeHtml(`<div style="color: #f59e0b; font-weight: bold;">${isRep ? "\uD83D\uDCE6 " : ""}⚠️ Divergente: ${state.lastScanned.desccompleta}</div>`);
    }
    renderHistory();
  });
  document.getElementById("btn-div-reset")?.addEventListener("click", finalizeActivity);
  document.getElementById("btn-predio-switch-yes")?.addEventListener("click", async () => {
    if (!state.atividade || !state.lastScanned)
      return;
    try {
      const newPredio = String(state.lastScanned.predio);
      const predios = state.atividade.predios || [
        state.atividade.predio
      ];
      const isNewBuilding = !predios.includes(newPredio);
      if (isNewBuilding) {
        predios.push(newPredio);
        state.atividade.predios = predios;
      }
      state.atividade.currentPredio = newPredio;
      if (isNewBuilding) {
        showLoader(true);
        const result = await apiCall(`/api/produtos/local?empresa=${state.atividade.empresa}&seqlocal=${state.atividade.seqlocal}&rua=${state.atividade.rua}&predio=${newPredio}`, {}, () => showReauthModal(true));
        showLoader(false);
        if (result.ok && Array.isArray(result.data) && result.data.length > 0) {
          const existingSeqs = new Set(state.expectedProducts.map((p) => p.seqproduto));
          const newProducts = result.data.filter((p) => !existingSeqs.has(p.seqproduto));
          state.expectedProducts.push(...newProducts);
        }
      }
      const isExpected = state.expectedProducts.some((p) => p.seqproduto === state.lastScanned.seqproduto);
      const newStatus = isExpected ? "OK" : "DIVERGENTE";
      let lastIdx = -1;
      for (let i = state.scannedProducts.length - 1;i >= 0; i--) {
        if (state.scannedProducts[i].seqproduto === state.lastScanned.seqproduto) {
          lastIdx = i;
          break;
        }
      }
      if (lastIdx >= 0) {
        state.scannedProducts[lastIdx].status = newStatus;
        state.scannedProducts[lastIdx].predio = newPredio;
      }
      saveState();
      showScreen("scanning");
      const feedback = document.getElementById("scan-feedback");
      if (feedback) {
        if (isExpected) {
          playBeep("success");
          feedback.innerHTML = sanitizeHtml(`<div style="color: #10b981; font-weight: bold;">✅ Prédio ${newPredio} agora é o prédio atual. Produto OK.</div>`);
        } else {
          playBeep("warning");
          feedback.innerHTML = sanitizeHtml(`<div style="color: #f59e0b; font-weight: bold;">⚠️ Prédio ${newPredio} agora é o prédio atual, porém produto não esperado!</div>`);
        }
      }
      state.lastScanned = null;
      renderHistory();
    } catch (e) {
      console.error("Error during building switch:", e);
      showScreen("scanning");
      const feedback = document.getElementById("scan-feedback");
      if (feedback)
        feedback.innerHTML = sanitizeHtml('<div style="color: #ef4444; font-weight: bold;">❌ Erro ao trocar de prédio</div>');
    }
  });
  document.getElementById("btn-predio-switch-no")?.addEventListener("click", () => {
    showScreen("scanning");
    const feedback = document.getElementById("scan-feedback");
    if (feedback && state.lastScanned) {
      feedback.innerHTML = sanitizeHtml(`<div style="color: #f59e0b; font-weight: bold;">⚠️ Divergente: ${state.lastScanned.desccompleta}</div>`);
    }
    renderHistory();
  });
  document.getElementById("btn-report-ok")?.addEventListener("click", () => {
    showScreen("start");
  });
  const openConsulta = () => {
    state.lastScanned = null;
    saveState();
    showScreen("consulta");
    let lojaNome = "";
    if (state.previousScreen === "scanning" && state.atividade) {
      lojaNome = `• ${state.atividade.empresaNome || `Loja ${state.atividade.empresa}`}`;
    } else {
      const empSelect = document.getElementById("start-empresa");
      if (empSelect?.selectedIndex >= 0) {
        const fullText = empSelect.options[empSelect.selectedIndex].text;
        lojaNome = `• ${fullText.split(" - ")[1] || fullText}`;
      }
    }
    const headerLoja = document.getElementById("consulta-header-loja");
    if (headerLoja)
      headerLoja.innerText = lojaNome;
    const input = document.getElementById("consulta-input");
    if (input) {
      input.value = "";
      input.focus();
    }
    document.getElementById("consulta-result")?.classList.add("hidden");
    document.getElementById("consulta-empty")?.classList.remove("hidden");
  };
  document.getElementById("btn-go-consulta")?.addEventListener("click", openConsulta);
  document.getElementById("btn-start-consulta")?.addEventListener("click", openConsulta);
  document.getElementById("btn-consulta-back")?.addEventListener("click", () => {
    const backTo = state.previousScreen || (state.atividade ? "scanning" : "start");
    showScreen(backTo === "consulta" ? "start" : backTo);
  });
  document.getElementById("form-consulta")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const input = document.getElementById("consulta-input");
    const code = input.value.trim();
    if (!code)
      return;
    let empresa = null;
    let seqlocal = null;
    let lojaNome = "";
    if (state.previousScreen === "scanning" && state.atividade) {
      empresa = state.atividade.empresa;
      seqlocal = state.atividade.seqlocal;
      lojaNome = state.atividade.empresaNome || `Loja ${empresa}`;
    } else {
      const empSelect = document.getElementById("start-empresa");
      const locSelect = document.getElementById("start-local");
      empresa = empSelect?.value;
      seqlocal = locSelect?.value;
      if (empSelect?.selectedIndex >= 0) {
        const fullText = empSelect.options[empSelect.selectedIndex].text;
        lojaNome = fullText.split(" - ")[1] || fullText;
      }
    }
    if (!empresa || !seqlocal) {
      alert("Selecione uma empresa e local primeiro");
      return;
    }
    showLoader(true);
    const { ok, data } = await apiCall(`/api/produtos/consulta/${code}?empresa=${empresa}&seqlocal=${seqlocal}`, {}, () => showReauthModal(true));
    showLoader(false);
    if (ok) {
      playBeep("success");
      document.getElementById("consulta-empty")?.classList.add("hidden");
      const resultEl = document.getElementById("consulta-result");
      if (resultEl)
        resultEl.classList.remove("hidden");
      const setVal = (id, val) => {
        const el = document.getElementById(id);
        if (el)
          el.innerText = String(val ?? "");
      };
      setVal("consulta-nome", data.desccompleta);
      setVal("consulta-loja-name", lojaNome);
      setVal("consulta-seq", data.seqproduto);
      setVal("consulta-marca", data.marca);
      setVal("consulta-estoque", data.estoque);
      setVal("consulta-dias", data.diasEstoque);
      setVal("consulta-mdv", data.mdv.toFixed(2).replace(".", ","));
      setVal("consulta-preco", data.precoVenda ? `R$ ${data.precoVenda.toFixed(2).replace(".", ",")}` : "N/A");
      setVal("consulta-entrada", data.dtaUltEntrada ? formatDate(data.dtaUltEntrada) : "N/A");
      setVal("consulta-venda", data.dtaUltVenda ? formatDate(data.dtaUltVenda) : "N/A");
      const codigosEl = document.getElementById("consulta-codigos");
      if (codigosEl && data.codigos) {
        codigosEl.innerHTML = sanitizeHtml(data.codigos.split("|").map((c) => `<span class="ean-badge">${c}</span>`).join(" "));
      }
      input.select();
    } else {
      playBeep("error");
      alert("Produto não encontrado");
      input.select();
    }
  });
});
