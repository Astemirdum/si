package ast_test

import (
	t "testing"

	"github.com/Astemirdum/si/internal/compiler"

	"github.com/stretchr/testify/suite"
)

type TinyProgramsTestSuite struct {
	compiler.Suite
}

func (suite *TinyProgramsTestSuite) TestRevLinkedList() {
	src := `
	type struct {
		i64 data,
		Node *next,
	} Node;

	i64 printf(i8 *fmt,... );
	i8* malloc(i64 size);
	
	
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
`

	expected := "Original list: 1 2 3 4 \nReversed list: 4 3 2 1 \n"
	suite.EqualProgramSi(src, expected)
}

func (suite *TinyProgramsTestSuite) TestBubbleSort() {
	src := `
i8* malloc(i64 size);
i8 free(i8* ptr);
i64 printf(i8 *fmt, ...);
i64 scanf(i8 *fmt, ...);

i64 swap(i64 *xp, i64 *yp) {
    i64 temp = *xp;
    *xp = *yp;
    *yp = temp;

    return 0;
}

i64 bubbleSort(i64* arr, i64 n) {
   i64 i;
   i64 j;

   for (i = 0; i < n; i++;) {
       for (j = 0; j < n-i-1; j++;){
           if (arr[j] > arr[j+1])
              swap(&arr[j], &arr[j+1]);
       }
   }
   return 0;
}

i64 printArray(i64* arr, i64 size) {
   i64 i = 0;
   while (i < size) {
       printf("%d ", arr[i]);
       i++;
   }
   printf("\n");
   return 0;
}

i64 main() {
    i64 n = 6; //sizeof(arr)/sizeof(arr[0]);
    i64* arr = (i64*)malloc(sizeof(i64) * n);
    arr[0] = 64;
    arr[1] = 32;
    arr[2] = 100;
    arr[3] = 0;
    arr[4] = -1;
    arr[5] = -1;

	printf("Unsorted array: \n");
	printArray(arr, n);
    
	bubbleSort(arr, n);
    printf("Sorted array: \n");
    printArray(arr, n);
	
    return 0;
}
`

	suite.EqualProgramSi(src, "Unsorted array: \n64 32 100 0 -1 -1 \nSorted array: \n-1 -1 0 32 64 100 \n")
}

func (suite *TinyProgramsTestSuite) TestFib() {
	src := `
i64 printf(i8 *fmt, ...);

i64 fib(i64 n) {
	if (n < 2) {
		return n;
	}

	return fib(n - 1) + fib(n - 2);
}

i64 main() {
	i64 n = 5;
	printf("fib(%d) = %d\n", n, fib(n));
	return 0;
}
`

	expected := `fib(5) = 5
`
	suite.EqualProgramSi(src, expected)
}

func (suite *TinyProgramsTestSuite) TestStrRev() {
	src := `
	i64 printf(i8 *fmt, ...);

	i8* strrev(i8 *s) {
		i8 *p = s;
		i8 *q = s;
		i8 tmp;

		while (*q != (i8)0) {
			q++;
		}

		q--;

		while (p < q) {
			tmp = *p;
			*p = *q;
			*q = tmp;
			p++;
			q--;
		}

		return s;
	}

	i64 main() {
		i8 *s = "hello";
		printf("%s", strrev(s));
		
		return 0;
	}
	`
	suite.EqualProgramSi(src, `olleh`)
}

func (suite *TinyProgramsTestSuite) TestBinSearch() {
	src := `
	i8* malloc(i64 size);
	i8 free(i8* ptr);
	i64 printf(i8 *fmt, ...);

	i64 binsearch(i64 x, i64* v, i64 n) {
		i64 low = 0;
		i64 high = n - 1;
		i64 mid;

		while (low <= high) {
			mid = (low + high) / 2;
			if (x < v[mid]) {
				high = mid - 1;
			} else if (x > v[mid]) {
				low = mid + 1;
			} else {
				return mid;
			}
		}

		return -1;
	}

	i64 main() {
		i64* v = (i64*)malloc(sizeof(i64) * 10);
		v[0] = 1;
		v[1] = 2;
		v[2] = 3;
		v[3] = 4;
		v[4] = 5;
		v[5] = 6;
		v[6] = 7;
		v[7] = 8;
		v[8] = 9;
		v[9] = 10;

		i64 x = 5;
		i64 n = 10;

		printf("%d", binsearch(x, v, n));

		return 0;
	}
	`
	suite.EqualProgramSi(src, "4")
}

func TestTinyProgramsTestSuite(t *t.T) {
	suite.Run(t, new(TinyProgramsTestSuite))
}
