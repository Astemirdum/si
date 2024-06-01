# Si-lang

## Intro

Si (Si-lang) is a strongly-typed, C-like programming language. It is written in Go and uses LLVM as a target backend.

tests
```bash
make test
```

compile
```bash
make compile f="example/fib.si"
```

## Examples

Si can link to standard C libraries and call functions from them. Here is an example of a program that calls the ```printf``` function from the standard C library.

Reversed LinkedList 
```c
i64 printf(i8 *fmt,... );
i8* malloc(i64 size);

type struct {
		i64 data,
		Node *next,
	} Node;


Node* reverseList(Node* head) {
    Node* prev = (Node*)NULL;
    Node* current = head;
    Node* next = (Node*)NULL;

    while (current != (Node*)NULL) {
        next = current->next;
        current->next = prev;
        prev = current;
        current = next;
    }
    return prev;
}

i64 printList(Node* node) {
    while (node != (Node*)NULL) {
        printf("%d ", node->data);
        node = node->next;
    }
    printf("\n");
    return 0;
}

Node* newNode(i64 data) {
    Node* node = (Node*)malloc(sizeof(Node));
    node->data = data;
    node->next = (Node*)NULL;
    return node;
}

i64 main() {
    Node* head = newNode(1);
    head->next = newNode(2);
    head->next->next = newNode(3);
    head->next->next->next = newNode(4);

    printf("Original list: ");
    printList(head);

    head = reverseList(head);

    printf("Reversed list: ");
    printList(head);

    return 0;
}
```

## Architecture

1. [alecthomas/participle](https://github.com/alecthomas/participle) is used in the ```parser``` package to parse the source code into an AST.
2. Transform() called on the AST and returns a new AST that is easier to work with in the next (code generation step) step.
3. In the ```ast``` package [llir/llvm](https://github.com/llir/llvm) is used for code generation. It is a Go package that generates easy to read, plain text LLVM IR.
4. We call ```llc``` to compile the LLVM IR into machine code.


## Useful Resources

### Code Generation
- [LLVM IR and Go](https://blog.gopheracademy.com/advent-2018/llvm-ir-and-go/) This is a good introduction to LLVM IR and how to use the llir/llvm package to generate LLVM IR in Go. This inspired me to start writing this compiler.

### C syntax
- [The syntax of C in Backus-Naur Form](https://cs.wmich.edu/~gupta/teaching/cs4850/sumII06/The%20syntax%20of%20C%20in%20Backus-Naur%20form.htm) This helped me a lot to understand the syntax of C.
- [Participle's MicroC](https://github.com/alecthomas/participle/tree/master/_examples/microc) This is a good example of how to use participle to parse C-like syntax.

### LLVM IR
- [LLVM IR reference](https://llvm.org/docs/LangRef.html) This is the official LLVM IR reference.
