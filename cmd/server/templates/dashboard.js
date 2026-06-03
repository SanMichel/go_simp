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

// src/client/admin/state.ts
var token = null;
var user = JSON.parse(localStorage.getItem("simp_user") || "null");
var selectedActivities = new Set;
var filterState = {
  username: [],
  empresa: [],
  rua: [],
  predio: [],
  impresso: [],
  id: []
};
var dateFilterState = { dataFimStart: "", dataFimEnd: "" };
var sortState = { column: null, direction: null };
var filterOptions = {
  username: [],
  empresa: [],
  rua: [],
  predio: [],
  impresso: [],
  id: []
};
var openDropdownColumn = null;
var currentModalItems = [];
var modalSort = {
  col: null,
  dir: null
};
function setToken(val) {
  token = val;
}
function setUser(val) {
  user = val;
}
function setFilterOptions(val) {
  filterOptions = val;
}
function setOpenDropdownColumn(val) {
  openDropdownColumn = val;
}
function setCurrentModalItems(val) {
  currentModalItems = val;
}
function setModalSort(val) {
  modalSort = val;
}
var COLUMN_LABELS = {
  username: "Usuário",
  empresa: "Empresa",
  rua: "Rua",
  predio: "Prédio",
  impresso: "Impresso",
  id: "ID",
  dataFim: "Data Fim"
};

// src/client/admin/helpers.ts
var _logoutCb = null;
function setLogoutCb(cb) {
  _logoutCb = cb;
}
async function api(path, options = {}) {
  return apiCall(path, options, () => _logoutCb?.());
}
function rolePt(role) {
  const map = {
    sysadmin: "Sysadmin",
    gerente: "Gerente",
    conferente: "Conferente"
  };
  return map[role] || role;
}
function fmtDate(iso) {
  if (!iso)
    return '<span style="color:var(--muted)">—</span>';
  try {
    return new Date(iso).toLocaleDateString("pt-BR");
  } catch {
    return iso;
  }
}
function fmtDateShort(dateStr) {
  if (!dateStr)
    return "";
  const parts = dateStr.split("-");
  if (parts.length === 3)
    return `${parts[2]}/${parts[1]}/${parts[0]}`;
  return dateStr;
}
function setMsg(el, text2, ok) {
  el.textContent = text2;
  el.className = `msg ${ok ? "msg-ok" : "msg-err"}`;
  if (ok)
    setTimeout(() => {
      el.textContent = "";
      el.className = "msg";
    }, 3000);
}
async function loadFilterOptions() {
  const result = await api("/api/dashboard/activities/filters");
  if (result.ok && result.data) {
    setFilterOptions(result.data);
  }
}
function buildQueryString() {
  const params = [];
  params.push("limit=100");
  const keys = ["username", "empresa", "rua", "predio", "impresso", "id"];
  for (const key of keys) {
    if (filterState[key] && filterState[key].length > 0) {
      params.push(`filter_${key}=${encodeURIComponent(filterState[key].join(","))}`);
    }
  }
  if (dateFilterState.dataFimStart) {
    params.push(`filter_dataFimStart=${encodeURIComponent(dateFilterState.dataFimStart)}`);
  }
  if (dateFilterState.dataFimEnd) {
    params.push(`filter_dataFimEnd=${encodeURIComponent(dateFilterState.dataFimEnd)}`);
  }
  if (sortState.column && sortState.direction) {
    params.push(`sort=${sortState.column}`);
    params.push(`order=${sortState.direction}`);
  }
  return params.join("&");
}
function updateFilterActiveClasses() {
  const ths = document.querySelectorAll(".th-filterable[data-column]");
  ths.forEach((th) => {
    const col = th.getAttribute("data-column");
    if (!col)
      return;
    if (col === "dataFim") {
      if (dateFilterState.dataFimStart || dateFilterState.dataFimEnd) {
        th.classList.add("filter-active");
      } else {
        th.classList.remove("filter-active");
      }
    } else if (filterState[col] && filterState[col].length > 0) {
      th.classList.add("filter-active");
    } else {
      th.classList.remove("filter-active");
    }
  });
}
function updateActiveFiltersBar() {
  const bar = document.getElementById("active-filters-bar");
  if (!bar)
    return;
  const keys = ["username", "empresa", "rua", "predio", "impresso", "id"];
  let hasAny = false;
  let html2 = '<span class="active-filter-label">Filtros:</span>';
  for (const key of keys) {
    if (filterState[key] && filterState[key].length > 0) {
      hasAny = true;
      for (const val of filterState[key]) {
        html2 += '<span class="active-filter-tag">' + (COLUMN_LABELS[key] || key) + ": " + escHtml(val) + ' <span class="tag-remove" data-filter-action="remove-value" data-filter-column="' + key + '" data-filter-value="' + escHtml(val) + '">&times;</span>' + "</span>";
      }
    }
  }
  if (dateFilterState.dataFimStart || dateFilterState.dataFimEnd) {
    hasAny = true;
    let dateLabel = "Data Fim: ";
    if (dateFilterState.dataFimStart && dateFilterState.dataFimEnd) {
      dateLabel += fmtDateShort(dateFilterState.dataFimStart) + " → " + fmtDateShort(dateFilterState.dataFimEnd);
    } else if (dateFilterState.dataFimStart) {
      dateLabel += `a partir de ${fmtDateShort(dateFilterState.dataFimStart)}`;
    } else {
      dateLabel += `até ${fmtDateShort(dateFilterState.dataFimEnd)}`;
    }
    html2 += '<span class="active-filter-tag">' + dateLabel + ' <span class="tag-remove" data-filter-action="clear-date">&times;</span>' + "</span>";
  }
  if (hasAny) {
    html2 += '<button class="btn-clear-filters" data-filter-action="clear-all">Limpar Filtros</button>';
    bar.innerHTML = sanitizeHtml(html2);
    bar.classList.remove("hidden");
  } else {
    bar.classList.add("hidden");
    bar.innerHTML = "";
  }
  updateFilterActiveClasses();
}
function closeFilterDropdown() {
  const existing = document.getElementById("filter-dropdown-active");
  if (existing) {
    const parent = existing.parentElement;
    existing.remove();
    if (parent) {
      const col = parent.getAttribute("data-column");
      if (col && (!filterState[col] || filterState[col].length === 0)) {
        parent.classList.remove("filter-active");
      }
    }
  }
  setOpenDropdownColumn(null);
}

// src/client/admin/activities.ts
async function loadActivities() {
  const tbody = document.getElementById("activities-tbody");
  if (!tbody)
    return;
  tbody.innerHTML = "";
  const loadingTr = document.createElement("tr");
  const loadingTd = document.createElement("td");
  loadingTd.colSpan = 9;
  const loadingDiv = document.createElement("div");
  loadingDiv.className = "empty";
  loadingDiv.textContent = "Carregando…";
  loadingTd.appendChild(loadingDiv);
  loadingTr.appendChild(loadingTd);
  tbody.appendChild(loadingTr);
  selectedActivities.clear();
  updateBulkPrintButton();
  const selectAllCheckbox = document.getElementById("select-all-activities");
  if (selectAllCheckbox)
    selectAllCheckbox.checked = false;
  if (filterOptions.username.length === 0 && filterOptions.id.length === 0) {
    await loadFilterOptions();
  }
  updateActiveFiltersBar();
  const qs = buildQueryString();
  const result = await api(`/api/dashboard/activities?${qs}`);
  if (!result.ok || !Array.isArray(result.data)) {
    tbody.innerHTML = "";
    const errTr = document.createElement("tr");
    const errTd = document.createElement("td");
    errTd.colSpan = 9;
    const errDiv = document.createElement("div");
    errDiv.className = "empty";
    errDiv.style.color = "var(--danger)";
    errDiv.textContent = "Erro ao carregar atividades.";
    errTd.appendChild(errDiv);
    errTr.appendChild(errTd);
    tbody.appendChild(errTr);
    return;
  }
  const data = result.data;
  const statTotalAct = document.getElementById("stat-total-act");
  if (statTotalAct)
    statTotalAct.textContent = data.length.toString();
  const statFinalizadas = document.getElementById("stat-finalizadas");
  if (statFinalizadas)
    statFinalizadas.textContent = data.filter((a) => a.dataFim).length.toString();
  const empresaSet = {};
  for (const act of data) {
    empresaSet[act.empresa] = true;
  }
  const statEmpresasAct = document.getElementById("stat-empresas-act");
  if (statEmpresasAct)
    statEmpresasAct.textContent = Object.keys(empresaSet).length.toString();
  if (data.length === 0) {
    tbody.innerHTML = "";
    const emptyTr = document.createElement("tr");
    const emptyTd = document.createElement("td");
    emptyTd.colSpan = 9;
    const emptyDiv = document.createElement("div");
    emptyDiv.className = "empty";
    emptyDiv.textContent = "Nenhuma atividade encontrada.";
    emptyTd.appendChild(emptyDiv);
    emptyTr.appendChild(emptyTd);
    tbody.appendChild(emptyTr);
    return;
  }
  tbody.innerHTML = "";
  for (const a of data) {
    const tr = buildActivityRow(a);
    tbody.appendChild(tr);
  }
}
function buildActivityRow(a) {
  const tr = document.createElement("tr");
  tr.setAttribute("data-activity-id", String(a.id));
  const cbCell = document.createElement("td");
  cbCell.className = "checkbox-col";
  const cb = document.createElement("input");
  cb.type = "checkbox";
  cb.className = "activity-checkbox";
  cb.setAttribute("data-activity-id", String(a.id));
  cbCell.appendChild(cb);
  tr.appendChild(cbCell);
  const idCell = document.createElement("td");
  idCell.style.color = "var(--muted)";
  idCell.style.fontSize = "12px";
  idCell.textContent = `#${a.id}`;
  tr.appendChild(idCell);
  const userCell = document.createElement("td");
  const userStrong = document.createElement("strong");
  userStrong.textContent = a.username || "—";
  userCell.appendChild(userStrong);
  tr.appendChild(userCell);
  const empCell = document.createElement("td");
  empCell.textContent = a.empresa || "—";
  tr.appendChild(empCell);
  const ruaCell = document.createElement("td");
  ruaCell.textContent = a.rua || "—";
  tr.appendChild(ruaCell);
  const predioCell = document.createElement("td");
  predioCell.textContent = a.predio || "—";
  tr.appendChild(predioCell);
  const dateCell = document.createElement("td");
  dateCell.style.color = "var(--muted)";
  dateCell.style.fontSize = "12px";
  dateCell.textContent = fmtDate(a.dataFim);
  tr.appendChild(dateCell);
  const imprCell = document.createElement("td");
  imprCell.style.textAlign = "center";
  imprCell.style.fontWeight = "bold";
  imprCell.style.color = a.impresso ? "#10b981" : "var(--muted)";
  imprCell.textContent = a.impresso ? "S" : "N";
  tr.appendChild(imprCell);
  const actionCell = document.createElement("td");
  const printBtn = document.createElement("button");
  printBtn.className = "btn btn-ghost btn-print";
  printBtn.setAttribute("data-activity-id", String(a.id));
  printBtn.setAttribute("data-action", "print");
  printBtn.title = "Imprimir";
  printBtn.textContent = "\uD83D\uDDA8️";
  actionCell.appendChild(printBtn);
  tr.appendChild(actionCell);
  return tr;
}
function toggleActivitySelection(id, checked) {
  if (checked) {
    selectedActivities.add(id);
  } else {
    selectedActivities.delete(id);
  }
  updateBulkPrintButton();
}
function toggleAllActivities(checked) {
  const checkboxes = document.querySelectorAll(".activity-checkbox");
  checkboxes.forEach((cb) => {
    cb.checked = checked;
    const id = parseInt(cb.getAttribute("data-activity-id") ?? "0", 10);
    if (checked)
      selectedActivities.add(id);
    else
      selectedActivities.delete(id);
  });
  updateBulkPrintButton();
}
function updateBulkPrintButton() {
  const btn = document.getElementById("btn-bulk-print");
  if (!btn)
    return;
  if (selectedActivities.size > 0) {
    btn.classList.remove("hidden");
    btn.textContent = `\uD83D\uDDA8️ Imprimir (${selectedActivities.size})`;
  } else {
    btn.classList.add("hidden");
  }
}

// src/client/admin/auth.ts
var _loadUsersCb = null;
function setLoadUsersCb(cb) {
  _loadUsersCb = cb;
}
var _allowedRoles = null;
function setAllowedRoles(roles) {
  _allowedRoles = roles;
}
function updateAuthState() {
  const loginScreen = document.getElementById("admin-login-screen");
  const appScreen = document.getElementById("app");
  const allowed = _allowedRoles || ["sysadmin"];
  if (!token || !user || !allowed.includes(user.role)) {
    loginScreen?.classList.remove("hidden");
    appScreen?.classList.add("hidden");
  } else {
    loginScreen?.classList.add("hidden");
    appScreen?.classList.remove("hidden");
    initDashboard();
  }
}
function initDashboard() {
  if (!user)
    return;
  const sidebarUsername = document.getElementById("sidebar-username");
  if (sidebarUsername)
    sidebarUsername.textContent = user.username;
  const sidebarRole = document.getElementById("sidebar-role");
  if (sidebarRole)
    sidebarRole.textContent = rolePt(user.role);
  const userAvatar = document.getElementById("user-avatar");
  if (userAvatar)
    userAvatar.textContent = user.username.charAt(0).toUpperCase();
  const userStatsRow = document.getElementById("user-stats-row");
  if (user.role !== "sysadmin" && userStatsRow) {
    userStatsRow.style.display = "none";
  }
  _loadUsersCb?.();
}
function logout() {
  apiCall("/api/auth/logout", { method: "POST" }).catch(() => {});
  setToken(null);
  setUser(null);
  localStorage.removeItem("simp_user");
  updateAuthState();
}

// src/client/admin/filters.ts
function handleColumnClick(column, event) {
  event.stopPropagation();
  if (openDropdownColumn === column) {
    closeFilterDropdown();
    return;
  }
  closeFilterDropdown();
  openFilterDropdown(column, event.currentTarget);
}
function handleSort(column, event) {
  event.stopPropagation();
  closeFilterDropdown();
  cycleSort(column);
}
function cycleSort(column) {
  if (sortState.column === column) {
    if (sortState.direction === "asc") {
      sortState.direction = "desc";
    } else if (sortState.direction === "desc") {
      sortState.column = null;
      sortState.direction = null;
    }
  } else {
    sortState.column = column;
    sortState.direction = "asc";
  }
  updateSortIndicators();
  loadActivities();
}
function updateSortIndicators() {
  const allCols = [
    "id",
    "username",
    "empresa",
    "rua",
    "predio",
    "dataFim",
    "impresso"
  ];
  for (const col of allCols) {
    const el = document.getElementById(`sort-${col}`);
    if (!el)
      continue;
    if (sortState.column === col) {
      el.classList.add("active");
      el.textContent = sortState.direction === "asc" ? "↑" : "↓";
    } else {
      el.classList.remove("active");
      el.textContent = "⇅";
    }
  }
}
function openFilterDropdown(column, thElement) {
  setOpenDropdownColumn(column);
  if (column === "dataFim") {
    openDateFilterDropdown(thElement);
    return;
  }
  const options = filterOptions[column] || [];
  const selected = filterState[column] || [];
  const dropdown = document.createElement("div");
  dropdown.className = "filter-dropdown";
  dropdown.id = "filter-dropdown-active";
  dropdown.setAttribute("data-filter-column", column);
  const searchHtml = '<div class="filter-search-wrap">' + '<input type="text" class="filter-search" placeholder="Pesquisar..." id="filter-search-input">' + "</div>";
  const actionsHtml = '<div class="filter-actions">' + '<button data-filter-action="select-all">Todos</button>' + '<button data-filter-action="select-none">Nenhum</button>' + '<button data-filter-action="sort">Ordenar</button>' + "</div>";
  let listHtml = '<div class="filter-list" id="filter-list-items">';
  for (const opt of options) {
    const isChecked = selected.indexOf(opt) === -1 ? "" : " checked";
    listHtml += '<label class="filter-item" data-value="' + escHtml(opt.toLowerCase()) + '">' + '<input type="checkbox" value="' + escHtml(opt) + '"' + isChecked + ' data-filter-column="' + column + '" data-filter-value="' + escHtml(opt) + '">' + '<span class="filter-item-label">' + escHtml(opt) + "</span>" + "</label>";
  }
  listHtml += "</div>";
  dropdown.innerHTML = sanitizeHtml(searchHtml + actionsHtml + listHtml);
  thElement.appendChild(dropdown);
  thElement.classList.add("filter-active");
  const searchInput = document.getElementById("filter-search-input");
  if (searchInput) {
    searchInput.addEventListener("input", function() {
      const term = this.value.toLowerCase();
      const items = document.querySelectorAll("#filter-list-items .filter-item");
      for (let j = 0;j < items.length; j++) {
        const val = items[j].getAttribute("data-value") || "";
        if (val.indexOf(term) !== -1) {
          items[j].classList.remove("hidden-by-search");
        } else {
          items[j].classList.add("hidden-by-search");
        }
      }
    });
    setTimeout(() => searchInput.focus(), 50);
  }
}
function openDateFilterDropdown(thElement) {
  const dropdown = document.createElement("div");
  dropdown.className = "filter-dropdown";
  dropdown.id = "filter-dropdown-active";
  dropdown.setAttribute("data-filter-column", "dataFim");
  const startVal = dateFilterState.dataFimStart || "";
  const endVal = dateFilterState.dataFimEnd || "";
  const html2 = '<div class="filter-date-wrap">' + '<div class="filter-date-field">' + "<label>De</label>" + '<input type="date" id="date-filter-start" lang="pt-BR" value="' + startVal + '">' + "</div>" + '<div class="filter-date-field">' + "<label>Até</label>" + '<input type="date" id="date-filter-end" lang="pt-BR" value="' + endVal + '">' + "</div>" + '<div class="filter-date-actions">' + '<button class="btn-apply" data-filter-action="apply-date">Aplicar</button>' + '<button class="btn-clear" data-filter-action="clear-date">Limpar</button>' + "</div>" + '<div class="filter-actions" style="border-top: 1px solid var(--border); border-bottom: none;">' + '<button data-filter-action="sort">Ordenar</button>' + "</div>" + "</div>";
  dropdown.innerHTML = sanitizeHtml(html2);
  thElement.appendChild(dropdown);
  thElement.classList.add("filter-active");
}
function applyDateFilter() {
  const startEl = document.getElementById("date-filter-start");
  const endEl = document.getElementById("date-filter-end");
  dateFilterState.dataFimStart = startEl ? startEl.value : "";
  dateFilterState.dataFimEnd = endEl ? endEl.value : "";
  closeFilterDropdown();
  updateActiveFiltersBar();
  loadActivities();
}
function clearDateFilter() {
  dateFilterState.dataFimStart = "";
  dateFilterState.dataFimEnd = "";
  const startEl = document.getElementById("date-filter-start");
  const endEl = document.getElementById("date-filter-end");
  if (startEl)
    startEl.value = "";
  if (endEl)
    endEl.value = "";
  updateActiveFiltersBar();
  loadActivities();
}
function handleFilterCheck(column, value, checked) {
  if (!filterState[column])
    filterState[column] = [];
  if (checked) {
    if (filterState[column].indexOf(value) === -1) {
      filterState[column].push(value);
    }
  } else {
    const idx = filterState[column].indexOf(value);
    if (idx !== -1)
      filterState[column].splice(idx, 1);
  }
  updateActiveFiltersBar();
  loadActivities();
}
function filterSelectAll(column) {
  filterState[column] = [];
  const checkboxes = document.querySelectorAll('#filter-list-items input[type="checkbox"]');
  checkboxes.forEach((cb) => {
    cb.checked = false;
  });
  updateActiveFiltersBar();
  loadActivities();
}
function filterSelectNone(column) {
  const options = filterOptions[column] || [];
  filterState[column] = options.slice();
  const checkboxes = document.querySelectorAll('#filter-list-items input[type="checkbox"]');
  checkboxes.forEach((cb) => {
    cb.checked = true;
  });
  updateActiveFiltersBar();
  loadActivities();
}
function cycleSortFromDropdown(column) {
  closeFilterDropdown();
  cycleSort(column);
}
function clearAllFilters() {
  const keys = ["username", "empresa", "rua", "predio", "impresso", "id"];
  for (const key of keys) {
    filterState[key] = [];
  }
  dateFilterState.dataFimStart = "";
  dateFilterState.dataFimEnd = "";
  updateActiveFiltersBar();
  updateFilterActiveClasses();
  loadActivities();
}
function removeFilterValue(column, value) {
  if (!filterState[column])
    return;
  const idx = filterState[column].indexOf(value);
  if (idx !== -1)
    filterState[column].splice(idx, 1);
  updateActiveFiltersBar();
  updateFilterActiveClasses();
  loadActivities();
}

// src/client/admin/modal.ts
async function showActivityDetails(id) {
  showLoader(true);
  const { ok, data } = await api(`/api/dashboard/activities/${id}`);
  showLoader(false);
  if (!ok) {
    alert(data.error || "Erro ao carregar detalhes da atividade.");
    return;
  }
  setCurrentModalItems(data.items || []);
  setModalSort({ col: null, dir: null });
  const modalTitle = document.getElementById("modal-title");
  if (modalTitle)
    modalTitle.textContent = `Detalhes da Atividade`;
  const modalSubtitle = document.getElementById("modal-subtitle");
  if (modalSubtitle)
    modalSubtitle.textContent = `#${data.id} — ${fmtDate(data.dataFim)}`;
  const modalUser = document.getElementById("modal-user");
  if (modalUser)
    modalUser.textContent = data.username || "—";
  const modalLocal = document.getElementById("modal-local");
  if (modalLocal)
    modalLocal.textContent = `Empresa ${data.empresa} / Rua ${data.rua} / Prédio ${data.predio}`;
  renderModalItems();
  updateModalSortIndicators();
  document.getElementById("modal-history")?.classList.remove("hidden");
}
function renderModalItems() {
  const tbody = document.getElementById("modal-history-tbody");
  if (!tbody)
    return;
  tbody.innerHTML = "";
  if (currentModalItems.length === 0) {
    const tr = document.createElement("tr");
    const td = document.createElement("td");
    td.colSpan = 7;
    const div = document.createElement("div");
    div.className = "empty";
    div.textContent = "Nenhum item verificado.";
    td.appendChild(div);
    tr.appendChild(td);
    tbody.appendChild(tr);
    return;
  }
  for (const item of currentModalItems) {
    const tr = document.createElement("tr");
    const nameTd = document.createElement("td");
    const nameDiv = document.createElement("div");
    nameDiv.style.fontWeight = "600";
    nameDiv.textContent = item.desccompleta || "Sem descrição";
    nameTd.appendChild(nameDiv);
    const seqDiv = document.createElement("div");
    seqDiv.style.fontSize = "11px";
    seqDiv.style.color = "var(--muted)";
    seqDiv.textContent = `Seq: ${item.seqproduto}`;
    nameTd.appendChild(seqDiv);
    tr.appendChild(nameTd);
    const statusTd = document.createElement("td");
    statusTd.style.textAlign = "center";
    const statusWrap = document.createElement("div");
    statusWrap.style.display = "flex";
    statusWrap.style.flexDirection = "column";
    statusWrap.style.alignItems = "center";
    statusWrap.style.gap = "2px";
    const badge = document.createElement("span");
    const statusClass = item.status === "OK" ? "badge-ok" : item.status === "ERRO" || item.status === "DIVERGENTE" ? "badge-warning" : "badge-error";
    badge.className = `badge ${statusClass}`;
    badge.textContent = item.status || "";
    statusWrap.appendChild(badge);
    if (item.reposicao) {
      const repBadge = document.createElement("span");
      repBadge.className = "badge";
      repBadge.style.background = "var(--warning)";
      repBadge.style.color = "#000";
      repBadge.style.fontSize = "10px";
      repBadge.textContent = "\uD83D\uDCE6 REPOSIÇÃO";
      statusWrap.appendChild(repBadge);
    }
    statusTd.appendChild(statusWrap);
    tr.appendChild(statusTd);
    const expectedAddr = item.expectedRua != null && item.expectedPredio != null ? `${item.expectedRua}/${item.expectedPredio}` : "N/A";
    const expTd = document.createElement("td");
    expTd.style.textAlign = "center";
    expTd.style.fontSize = "12px";
    expTd.style.color = "var(--muted)";
    expTd.textContent = expectedAddr;
    tr.appendChild(expTd);
    const readAddr = item.rua != null && item.predio != null ? `${item.rua}/${item.predio}` : "N/A";
    const readTd = document.createElement("td");
    readTd.style.textAlign = "center";
    readTd.style.fontSize = "12px";
    readTd.style.fontWeight = "600";
    readTd.textContent = readAddr;
    tr.appendChild(readTd);
    const estTd = document.createElement("td");
    estTd.style.textAlign = "right";
    estTd.style.fontWeight = "600";
    estTd.textContent = item.estoque != null ? String(item.estoque) : "";
    tr.appendChild(estTd);
    const mdv = item.mdv !== null ? item.mdv : "—";
    const mdvTd = document.createElement("td");
    mdvTd.style.textAlign = "right";
    mdvTd.style.fontSize = "12px";
    mdvTd.textContent = mdv;
    tr.appendChild(mdvTd);
    const ddv = item.ddv !== null ? item.ddv : "—";
    const ddvTd = document.createElement("td");
    ddvTd.style.textAlign = "right";
    ddvTd.style.fontSize = "12px";
    ddvTd.style.fontWeight = "600";
    ddvTd.textContent = ddv;
    tr.appendChild(ddvTd);
    tbody.appendChild(tr);
  }
}
function handleModalSort(col) {
  if (modalSort.col === col) {
    modalSort.dir = modalSort.dir === "asc" ? "desc" : "asc";
  } else {
    modalSort.col = col;
    modalSort.dir = "asc";
  }
  currentModalItems.sort((a, b) => {
    let valA, valB;
    if (col === "produto") {
      valA = a.desccompleta || "";
      valB = b.desccompleta || "";
    } else if (col === "status") {
      valA = a.status || "";
      valB = b.status || "";
    } else if (col === "esperado") {
      valA = `${a.expectedRua || ""}/${a.expectedPredio || ""}`;
      valB = `${b.expectedRua || ""}/${b.expectedPredio || ""}`;
    } else if (col === "lido") {
      valA = `${a.rua || ""}/${a.predio || ""}`;
      valB = `${b.rua || ""}/${b.predio || ""}`;
    } else if (col === "estoque" || col === "mdv" || col === "ddv") {
      const itemA = a;
      const itemB = b;
      valA = itemA[col] ?? 0;
      valB = itemB[col] ?? 0;
    } else {
      valA = 0;
      valB = 0;
    }
    if (typeof valA === "string" && typeof valB === "string") {
      return modalSort.dir === "asc" ? valA.localeCompare(valB, undefined, { numeric: true }) : valB.localeCompare(valA, undefined, { numeric: true });
    }
    return modalSort.dir === "asc" ? valA - valB : valB - valA;
  });
  renderModalItems();
  updateModalSortIndicators();
}
function updateModalSortIndicators() {
  const allCols = [
    "produto",
    "status",
    "esperado",
    "lido",
    "estoque",
    "mdv",
    "ddv"
  ];
  for (const col of allCols) {
    const el = document.getElementById(`modal-sort-${col}`);
    if (!el)
      continue;
    if (modalSort.col === col) {
      el.classList.add("active");
      el.textContent = modalSort.dir === "asc" ? "↑" : "↓";
    } else {
      el.classList.remove("active");
      el.textContent = "⇅";
    }
  }
}
function closeModal() {
  document.getElementById("modal-history")?.classList.add("hidden");
}

// src/client/admin/printing.ts
async function handleBulkPrint() {
  if (selectedActivities.size === 0)
    return;
  showLoader(true);
  const ids = Array.from(selectedActivities);
  const { ok, data } = await api("/api/dashboard/activities/bulk", {
    method: "POST",
    body: JSON.stringify({ ids })
  });
  showLoader(false);
  if (!ok) {
    alert(data.error || "Erro ao carregar dados para impressão em lote.");
    return;
  }
  printMultipleActivities(data);
}
function printMultipleActivities(activities) {
  const printWindow = window.open("", "_blank");
  if (!printWindow) {
    alert("Bloqueador de pop-up impediu a impressão.");
    return;
  }
  let allReportsHtml = "";
  activities.forEach((activity, index) => {
    allReportsHtml += `<div class="activity-container">${generateActivityReportHtml(activity)}</div>`;
    if (index < activities.length - 1) {
      allReportsHtml += '<hr class="report-divider">';
    }
  });
  const finalHtml = `
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="UTF-8">
        <title>SIMP - Relatórios de Impressão</title>
        ${getReportStyles()}
    </head>
    <body>
        ${allReportsHtml}
    </body>
    </html>
    `;
  printWindow.document.write(finalHtml);
  printWindow.document.close();
  printWindow.focus();
  setTimeout(() => {
    printWindow.print();
    printWindow.close();
    const ids = activities.map((a) => a.id);
    api(`/api/dashboard/activities/bulk/print`, {
      method: "PATCH",
      body: JSON.stringify({ ids })
    }).then(() => loadActivities());
  }, 250);
}
function generateActivityReportHtml(data) {
  const itemsToPrint = (data.items || []).filter((item) => item.status === "RUPTURA" || item.reposicao).sort((a, b) => {
    const nameA = (a.desccompleta || "").toUpperCase();
    const nameB = (b.desccompleta || "").toUpperCase();
    if (nameA < nameB)
      return -1;
    if (nameA > nameB)
      return 1;
    return 0;
  });
  const rows = itemsToPrint.map((item) => {
    const expectedAddr = item.expectedRua != null && item.expectedPredio != null ? `${item.expectedRua}/${item.expectedPredio}` : "N/A";
    const formatNum = (v) => v !== null && v !== undefined ? v.toLocaleString("pt-BR", {
      minimumFractionDigits: 0,
      maximumFractionDigits: 2
    }) : "—";
    const mdv = formatNum(item.mdv);
    const ddv = formatNum(item.ddv);
    const dtOpts = {
      day: "2-digit",
      month: "2-digit",
      year: "2-digit"
    };
    const dtEntrada = item.dtaultentrada ? new Date(item.dtaultentrada).toLocaleDateString("pt-BR", dtOpts) : "—";
    let daysSinceLastSale = "—";
    if (item.dtaultvenda) {
      const lastSale = new Date(item.dtaultvenda);
      const today = new Date;
      const diffTime = Math.abs(today.getTime() - lastSale.getTime());
      daysSinceLastSale = Math.floor(diffTime / (1000 * 60 * 60 * 24));
    }
    return `
        <tr>
            <td class="col-seq">${item.seqproduto}</td>
            <td>
                ${item.desccompleta || "Sem descrição"}
                ${item.reposicao ? '<span style="margin-left:4px;" title="Reposição">\uD83D\uDCE6</span>' : ""}
            </td>
            <td class="col-addr">${expectedAddr}</td>
            <td style="text-align:right;">${item.estoque}</td>
            <td style="text-align:right;">${mdv}</td>
            <td style="text-align:right;">${ddv}</td>
            <td class="col-date">${dtEntrada}</td>
            <td class="col-date">${daysSinceLastSale}</td>
        </tr>
        `;
  }).join("");
  return `
        <div class="report-header">
            <div class="header-left">
                <h1>Relatório de Atividade #${data.id}</h1>
                <div class="subtitle">Gerado em ${new Date().toLocaleString("pt-BR")}</div>
            </div>

            <div class="header-right info">
                <div class="info-row">
                    <div>
                        <span><span class="info-label">Usuário:</span> ${data.username || "—"}</span>
                        <span><span class="info-label">Empresa:</span> ${data.empresa}</span>
                    </div>
                    <div>
                        <span><span class="info-label">Rua:</span> ${data.rua}</span>
                        <span><span class="info-label">Prédio:</span> ${data.predio}</span>
                        <span><span class="info-label">Data:</span> ${fmtDate(data.dataFim)}</span>
                    </div>
                </div>
            </div>
        </div>

        <table>
            <thead>
                <tr>
                    <th class="col-seq">Cod.</th>
                    <th>Descrição</th>
                    <th class="col-addr">End. Esp.</th>
                    <th style="text-align:right; width: 40px;">Estq</th>
                    <th style="text-align:right; width: 45px;">MDV</th>
                    <th style="text-align:right; width: 45px;">DDV</th>
                    <th class="col-date">Últ. Ent.</th>
                    <th class="col-date">Dias s/ Vnd.</th>
                </tr>
            </thead>
            <tbody>
                ${rows.length > 0 ? rows : '<tr><td colspan="8" style="text-align:center; padding: 20px;">Nenhuma ruptura ou reposição encontrada nesta atividade.</td></tr>'}
            </tbody>
        </table>

        <div class="footer">
            SIMP - Sistema de Monitoramento de Prateleiras
        </div>
    `;
}
function getReportStyles() {
  return `
        <style>
            * { box-sizing: border-box; margin: 0; padding: 0; }
            body { font-family: Arial, sans-serif; font-size: 11px; color: #000; padding: 20px; }
            h1 { font-size: 18px; margin-bottom: 4px; }
            .subtitle { color: #666; font-size: 12px; }
            .report-header {
                display: flex;
                justify-content: space-between;
                align-items: flex-start;
                margin-bottom: 15px;
            }
            .header-right { text-align: right; }
            .info { margin-bottom: 0; }
            .info-row { display: flex; flex-direction: column; align-items: flex-end; gap: 2px; }
            .info-row > div { display: flex; gap: 15px; }
            .info-label { font-weight: bold; color: #444; }
            table { width: 100%; border-collapse: collapse; margin-top: 10px; table-layout: auto; }
            th, td { border: 1px solid #ccc; padding: 4px 6px; text-align: left; word-wrap: break-word; }
            th { background: #f5f5f5; font-weight: bold; }
            tr { page-break-inside: avoid; }
            .footer { margin-top: 20px; font-size: 10px; color: #888; text-align: center; }
            .col-seq { width: 40px; }
            .col-addr { width: 60px; text-align: center; }
            .col-date { width: 65px; text-align: center; }
            .activity-container { margin-bottom: 20px; }
            .report-divider {
                border: none;
                border-top: 2px dashed #000;
                margin: 40px 0;
                page-break-inside: avoid;
            }
            @media print {
                body { padding: 0; }
                .report-divider { border-top-color: #aaa; }
            }
        </style>
    `;
}
async function printActivity(id) {
  showLoader(true);
  const { ok, data } = await api(`/api/dashboard/activities/${id}`);
  showLoader(false);
  if (!ok) {
    alert(data.error || "Erro ao carregar dados para impressão.");
    return;
  }
  printMultipleActivities([data]);
}

// src/client/dashboard/index.ts
setLogoutCb(logout);
window.filterSelectAll = filterSelectAll;
window.filterSelectNone = filterSelectNone;
window.cycleSortFromDropdown = cycleSortFromDropdown;
window.removeFilterValue = removeFilterValue;
window.clearAllFilters = clearAllFilters;
window.applyDateFilter = applyDateFilter;
window.clearDateFilter = clearDateFilter;
window.handleFilterCheck = handleFilterCheck;
window.cycleSort = cycleSort;
window.handleSort = handleSort;
window.handleColumnClick = handleColumnClick;
window.printActivity = printActivity;
window.showActivityDetails = showActivityDetails;
window.toggleActivitySelection = toggleActivitySelection;
window.toggleAllActivities = toggleAllActivities;
window.handleBulkPrint = handleBulkPrint;
window.handleModalSort = handleModalSort;
window.closeModal = closeModal;
window.loadActivities = loadActivities;
document.getElementById("form-admin-login")?.addEventListener("submit", async (e) => {
  e.preventDefault();
  showLoader(true);
  const errEl = document.getElementById("admin-login-error");
  errEl?.classList.add("hidden");
  const username = document.getElementById("admin-username").value;
  const password = document.getElementById("admin-password").value;
  const res = await apiCall("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password })
  });
  showLoader(false);
  if (res.ok) {
    if (!["sysadmin", "gerente"].includes(res.data.user.role)) {
      if (errEl) {
        errEl.textContent = "Acesso restrito";
        errEl.classList.remove("hidden");
      }
      return;
    }
    setToken(res.data.token);
    setUser(res.data.user);
    localStorage.setItem("simp_user", JSON.stringify(user));
    updateAuthState();
  } else if (errEl) {
    errEl.textContent = res.data.error || "Erro ao logar";
    errEl.classList.remove("hidden");
  }
});
document.getElementById("btn-logout")?.addEventListener("click", logout);
document.addEventListener("click", (e) => {
  if (openDropdownColumn && !e.target.closest(".filter-dropdown") && !e.target.closest(".th-filterable")) {
    closeFilterDropdown();
  }
});
document.addEventListener("click", (e) => {
  const filterBtn = e.target.closest("[data-filter-action]");
  if (!filterBtn)
    return;
  const dropdown = filterBtn.closest(".filter-dropdown");
  const column = dropdown?.getAttribute("data-filter-column") || "";
  const action = filterBtn.getAttribute("data-filter-action");
  if (action === "select-all")
    filterSelectAll(column);
  else if (action === "select-none")
    filterSelectNone(column);
  else if (action === "sort")
    cycleSortFromDropdown(column);
  else if (action === "apply-date")
    applyDateFilter();
  else if (action === "clear-date")
    clearDateFilter();
  else if (action === "clear-all")
    clearAllFilters();
  else if (action === "remove-value") {
    const val = filterBtn.getAttribute("data-filter-value") || "";
    const col = filterBtn.getAttribute("data-filter-column") || "";
    removeFilterValue(col, val);
  }
});
document.addEventListener("change", (e) => {
  const cb = e.target.closest("input[data-filter-column][data-filter-value]");
  if (!cb)
    return;
  const column = cb.getAttribute("data-filter-column") || "";
  const value = cb.getAttribute("data-filter-value") || "";
  handleFilterCheck(column, value, cb.checked);
});
document.getElementById("activities-tbody")?.addEventListener("click", (e) => {
  const printBtn = e.target.closest('button[data-action="print"]');
  if (printBtn) {
    const id = Number(printBtn.getAttribute("data-activity-id"));
    if (id)
      printActivity(id);
    return;
  }
  const row = e.target.closest("tr[data-activity-id]");
  if (row) {
    const id = Number(row.getAttribute("data-activity-id"));
    if (id)
      showActivityDetails(id);
  }
});
document.getElementById("activities-tbody")?.addEventListener("change", (e) => {
  const cb = e.target.closest(".activity-checkbox");
  if (!cb)
    return;
  const id = Number(cb.getAttribute("data-activity-id"));
  if (id)
    toggleActivitySelection(id, cb.checked);
});
document.addEventListener("DOMContentLoaded", async () => {
  const sessionRes = await apiCall("/api/auth/me", {}, () => {});
  if (sessionRes.ok && sessionRes.data?.user && ["sysadmin", "gerente"].includes(sessionRes.data.user.role)) {
    setToken("cookie");
    setUser(sessionRes.data.user);
    localStorage.setItem("simp_user", JSON.stringify(user));
  }
  setAllowedRoles(["sysadmin", "gerente"]);
  updateAuthState();
  loadFilterOptions();
  loadActivities();
});
