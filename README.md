# inspecthtml-go

Parse HTML and capture metadata about byte offsets.

* Reference byte and line+column offsets of nodes and element attributes.
* Wraps the standard [`golang.org/x/net/html`](https://pkg.go.dev/golang.org/x/net/html) implementation.

This is implemented by pre-tokenizing the input stream to inject offset metadata before forwarding it to `html.Parse` and then cleaning up smuggled metadata from the resulting tree. The returned result is a standard `*html.Node` result, practically equivalent to an `html.Parse`, and a service for retrieving offset metadata for given a `*html.Node`.

## Usage

Import the module and refer to the code's documentation ([pkg.go.dev](https://pkg.go.dev/github.com/dpb587/inspecthtml-go/inspecthtml)).

```go
import "github.com/dpb587/inspecthtml-go/inspecthtml"
```

Some sample use cases and starter snippets can be found in the [`examples` directory](examples).

<details><summary><code>examples$ go run ./<strong>parse-dump</strong> <<<'<strong>&lt;p class="headline"&gt;&lt;strong&gt;hello&lt;/strong&gt;&lt;br data-example /&gt;world&lt;!-- end--&gt;</strong>'</code></summary>

```
  <html>
    <head>
    </head>
    <body>
      // StartTagToken=L1C1:L1C21;0x0:0x14 OuterOffsets=L1C1:L2C1;0x0:0x4e InnerOffsets=L1C21:L2C1;0x14:0x4e
      <p
        // Attr KeyOffsets=L1C4:L1C9;0x3:0x8 ValueOffsets=L1C10:L1C20;0x9:0x13
        class="headline"
      >
        // StartTagToken=L1C21:L1C29;0x14:0x1c OuterOffsets=L1C21:L1C43;0x14:0x2a InnerOffsets=L1C29:L1C34;0x1c:0x21
        <strong>
          // TextToken=L1C29:L1C34;0x1c:0x21
          hello
        // EndTagToken=L1C34:L1C43;0x21:0x2a
        </strong>
        // StartTagToken=L1C43:L1C62;0x2a:0x3d OuterOffsets=L1C43:L1C62;0x2a:0x3d SelfClosing
        <br
          // Attr KeyOffsets=L1C47:L1C59;0x2e:0x3a
          data-example=""
        >
        </br>
        // TextToken=L1C62:L1C67;0x3d:0x42
        world
        // CommentToken=L1C67:L1C78;0x42:0x4d
        <!-- end-->
        // TextToken=L1C78:L2C1;0x4d:0x4e
        

      // EndTagToken=L2C1:L2C1;0x4e:0x4e
      </p>
    </body>
  </html>
```

</details>

More complex usage can be seen from importers like [rdfkit-go](https://github.com/dpb587/rdfkit-go).

## Parser

Given an `io.Reader`, parse and return a standard `*html.Node` as well as the resulting metadata.

```go
parsedNode, parsedMetadata, err := inspecthtml.Parse(os.Stdin)
```

For any node of interest, retrieve it from the metadata provider.

```go
nodeMetadata, hasNodeMetadata := parsedMetadata.GetNodeMetadata(node)
```

Always check that node or attribute metadata is available before accessing it. Specifically, keep in mind the following:

* The DOM Processor may inject elements or re-parent nodes to create a compliant HTML5 DOM Tree. Injected elements will not have any metadata since they were not present in source. Re-parented nodes may be siblings in the DOM, but have non-sequential source offsets.
* The DOM Processor will close unclosed elements. In this case, the metadata will use a logical end tag of zero length based on the relative position of the next element.
* Node attributes may not have an offset for its value if there was no value in the source.
* Although unlikely, it is technically possible for a node attribute to be missing metadata due to the implementation method. In this case, it will be `nil` (and a test case for fixing it would be helpful).
* A document parsed by both `html.Parse` and `inspecthtml.Parse` is expected to be equivalent after calling `html.Render` (the DOM trees may not be exactly equivalent). Known exceptions are:
  * All `style` tags contain data as unparsed text (vs `html` which parses the data in foreign elements, namely SVG and MathML, and may produce additional text or comment nodes).

## Notes

* This approach is a bit hacky and inefficient due to double tokenization steps and additional, injected placeholder tokens that must be processed. When getting started, at least, this seemed like a practical alternative to: maintain an indefinite fork with embedded offset range tracking; or reimplement the complex HTML5 processor specification. There are a couple spots where Go's tokenization/parser behaviors leak into this for the sake of accurate offsets tracking.
* [go#34302](https://github.com/golang/go/issues/34302) describes a feature proposal for a subset of this behavior, but has not seen any recent activity.

## License

[MIT License](LICENSE)
