package ast_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Astemirdum/si/internal/compiler"
	"github.com/stretchr/testify/suite"
)

type SrcTestSuite struct {
	compiler.Suite
}

func TestSrcTestSuite(t *testing.T) {
	suite.Run(t, new(SrcTestSuite))
}

func (suite *SrcTestSuite) SetupTest() {
	suite.T().Parallel()
}

func (suite *SrcTestSuite) TestBasic() {
	src := `
i64 printf(i8 *fmt,... );

i64 main() {
	printf("10");
	return 0;
}
`
	suite.EqualProgramSi(src, `10`)
}

func (suite *SrcTestSuite) TestBasicC() {
	src := `
#include <stdio.h>

int main() {
	printf("10");
	return 0;
}
`
	suite.EqualProgramC(src, `10`)
}

func (suite *SrcTestSuite) TestBasicInline() {
	suite.EqualExprSi("i64 t = 10;", `"%d", t`, "10")
}

func (suite *SrcTestSuite) TestMathInline() {
	suite.EqualExprSi("f64 t = round(10.6);", `"%.2f", t`, "11.00", compiler.Declare("f64 round(f64 arg);"))
}

func (suite *SrcTestSuite) TestBasicInlineC() {
	suite.EqualExprC("int t = 10;", `"%d", t`, "10")
}

func (suite *SrcTestSuite) TestMathInlineC() {
	suite.EqualExprC("float t = roundf(10.6f);", `"%.2f", t`, "11.00", compiler.Header("math.h"))
}

func (suite *SrcTestSuite) TestDoublePrefixC() {
	suite.ErrorClangExprC(`char *s = "hello"; ++(++s);`, "expression is not assignable")

	suite.EqualExprC(`char *s = "hello"; char *s0 = ++s; char *s1 = ++s;`, `"%s,%s", s0, s1`, "ello,llo")
	suite.EqualExprSi(`i8 *s = "hello";`, `"%s,%s,%s", ++s, ++(++s), s`, "ello,lo,lo")

	suite.EqualExprC(`char *s = "hello"; char *p = s;`, `"%s,%s", p + 2, s`, "llo,hello")
	suite.EqualExprSi(`i8 *s = "hello"; i8 *p = s;`, `"%s,%s", p + 2, s`, "llo,hello")

	// https://stackoverflow.com/questions/381542/with-arrays-why-is-it-the-case-that-a5-5a
	suite.EqualExprC(`char *s = "12345"; char p = s[2];`, `"%c", p`, "3")
	suite.EqualExprSi(`i8 *s = "12345"; i8 p = s[2];`, `"%c", p`, "3")

	suite.EqualExprSi(`i8 s = 'B';`, `"%c", s`, "B")
}

func (suite *SrcTestSuite) Test0() {
	suite.EqualExprC(`int a = 10; int *b = &a;`, `"%d", *b`, "10")
	suite.EqualExprSi(`i64 a = 10; i64 *b = &a;`, `"%d", *b`, "10")
}

func (suite *SrcTestSuite) Test1() {
	suite.EqualExprC(`char *s = "12345";`, `"%c", *(s + 2)`, "3")
	suite.EqualExprSi(`i8 *s = "12345";`, `"%c", *(s + 2)`, "3")

	suite.EqualExprC(`char *s = "12345";`, `"%s", &s[2]`, "345")
	suite.EqualExprSi(`i8 *s = "12345";`, `"%s", &s[2]`, "345")

	suite.EqualExprC(`char *s = malloc(10); memset(s, 'A', 6); s[2] = 'X'; s[5] = 0;`, `"%s-", s`, "AAXAA-", compiler.Header("stdlib.h"), compiler.Header("string.h"))
	suite.EqualExprSi(`i8 *s = malloc(10); memset(s, 'A', 6); s[2] = 'X'; s[5] = (i8)0;`, `"%s-", s`, "AAXAA-",
		compiler.DeclareMalloc(),
	)

	suite.EqualExprSi(`i8 *s = "12345"; *(s + 1) = 'C';`, `"%s", s`, "1C345")
	suite.EqualExprSi(`i8 *s = "12345"; i8 *p = s; *(++p) = 'F'; p[2] = 'X';`, `"%s", s`, "1F3X5")
}

func (suite *SrcTestSuite) Test2() {
	suite.EqualExprC(`char *s = "12345"; char **p = &s;`, `"%c", *(*p + 2)`, "3")
	suite.EqualExprSi(`i8 *s = "12345"; i8 **p = &s;`, `"%c", *(*p + 2)`, "3")

	suite.EqualExprC(`char *s = "12345"; char **p = malloc(sizeof(char*) * 3); p[1] = s;`, `"%c,%c", *(p[1] + 2), p[1][1]`, "3,2", compiler.Header("stdlib.h"))
	suite.EqualExprC(`char *s = "12345"; char **p = malloc(sizeof(char*) * 3); p[1] = s;`, `"%c,%c", *(p[1] + 2),(p[1])[1]`, "3,2", compiler.Header("stdlib.h"))
	suite.EqualExprSi(`i8 *s = "12345"; i8 **p = malloc(sizeof(i8*) * 3); p[1] = s;`, `"%c,%c", *(p[1] + 2), p[1][1]`, "3,2",
		compiler.Declare("i8** malloc(i64 size);"),
	)
	suite.EqualExprSi(`i8 *s = "12345"; i8 **p = malloc(sizeof(i8*) * 3); p[1] = s;`, `"%c,%c", *(p[1] + 2), (p[1])[1]`, "3,2",
		compiler.Declare("i8** malloc(i64 size);"),
	)
}

func (suite *SrcTestSuite) TestFn() {
	si := `

i64 printf(i8 *fmt,... );

void fn(i64 x) {
	printf("%d", x);
	return ;
}

i64 main() {
	i64 x = 10; 
	fn(x);
	return 0;
}
`
	suite.EqualProgramSi(si, "10")
}

func (suite *SrcTestSuite) TestMallocString() {
	si := `
i8* malloc(i64 size);
i8 free(i8* ptr);
i8* memset(i8* ptr, i8 val, i64 size);
i64 printf(i8 *fmt,... );

i64 main() {
	i8 *s = malloc(6);
	memset(s, 'h', 1);
	memset(s + 1, 'e', 1);
	memset(s + 2, 'l', 1);
	memset(s + 3, 'l', 1);
	memset(s + 4, 'o', 1);
	memset(s + 5, (i8)0, 1);
	printf("%s-", &s[2]);
	free(s);
	return 0;
}
`
	suite.EqualProgramSi(si, "llo-")
}

func (suite *SrcTestSuite) TestEmptyString() {
	suite.EqualExprSi(`i8 *a = "";`, `"-%s-", a`, "--")
	suite.EqualExprSi(`i8 *a = "a";`, `"-%s-", a`, "-a-")
}

func (suite *SrcTestSuite) TestAlias0a() {
	suite.ErrorGenerateExprSi(`hello a = (hello)10; bool d = a == 10;`, "incompatible types hello and i64", compiler.Declare("type i64 hello;"))
	suite.EqualExprSi(`hello a = (hello)10; bool d = a == (hello)10;`, `"%d", d`, "1", compiler.Declare("type i64 hello;"))
	suite.EqualExprSi(`i64 f = 10; hello g = (hello)10; bool h = f == (i64)g;`, `"%d", h`, "1", compiler.Declare("type i64 hello;"), compiler.Basename("first"))
}

func (suite *SrcTestSuite) TestAlias0b() {
	suite.EqualExprSi(`hello a = (hello)20; i8 b = (i8)a; i64 c = (i64)a;`, `"%d,%d,%d", a, b, c`, "20,20,20", compiler.Declare("type i32 hello;"), compiler.Basename("second"))

}

func (suite *SrcTestSuite) TestAlias1() {
	si := `
	type i64 hello;
	type hello world;
	type world* myptr;
	type bool bb;

	i64 printf(i8 *fmt,... );

	i64 main(hello b) {
		hello a = (hello)10;
		myptr c = (myptr)&a;
		bool d = ((i64)(*c) == 10);
		bb e = (bb)d;
		printf("%d-%c", c[0], e);
		return 0;
	}
	`
	suite.EqualProgramSi(si, "10-\x01")
}

func (suite *SrcTestSuite) TestWhile() {
	src := `
	i64 printf(i8 *fmt,... );

	i64 main() {
		i64 a = 10;

		while (a-- > 0) {
			printf("%d\n", a);
		}

		return 0;
	}
	`
	suite.EqualProgramSi(src, "9\n8\n7\n6\n5\n4\n3\n2\n1\n0\n")
}

func (suite *SrcTestSuite) TestFor() {

	suite.T().Run("For Expr", func(t *testing.T) {
		src := `
	i64 printf(i8 *fmt, ...);
	
	i64 main() {
		i64 x;

		for (x = 0; x < 10; x++;) {
			printf("%d\n", x);
		}

		return 0;
	}
`
		suite.EqualProgramSi(src, "0\n1\n2\n3\n4\n5\n6\n7\n8\n9\n")
	})

	suite.T().Run("For Assign", func(t *testing.T) {
		src := `
	i64 printf(i8 *fmt, ...);
	
	i64 main() {
		i64 x;

		for (x = 0; x < 10; x = x + 1;) {
			printf("%d\n", x);
		}

		return 0;
	}
`
		suite.EqualProgramSi(src, "0\n1\n2\n3\n4\n5\n6\n7\n8\n9\n")
	})
}

func (suite *SrcTestSuite) TestContinue() {
	src := `
	i64 printf(i8 *fmt, ...);
	
	i64 main() {
		i64 x = 0;

		while (x < 10) {
			x++;
			if (x < 5) {
				continue;
			}
			printf("%d\n", x);
		}

		return 0;
	}
	`
	suite.EqualProgramSi(src, "5\n6\n7\n8\n9\n10\n")
}

func (suite *SrcTestSuite) TestBreak() {
	suite.T().Run("while", func(t *testing.T) {
		src := `
	i64 printf(i8 *fmt, ...);
	
	i64 main() {
		i64 x = 0;

		while (x < 10) {
			if (x == 5) {
				break;
			}
			printf("%d\n", x);
			x++;
		}

		return 0;
	}
	`
		suite.EqualProgramSi(src, "0\n1\n2\n3\n4\n")
	})

	suite.T().Run("for", func(t *testing.T) {
		src := `
	i64 printf(i8 *fmt, ...);
	
	i64 main() {
		i64 x;

		for (x = 0; x < 10; x++;) {
			if (x == 5) {
				break;
			}
			printf("%d\n", x);
		}

		return 0;
	}
	`
		_ = src
		suite.EqualProgramSi(src, "0\n1\n2\n3\n4\n")
	})
}

func (suite *SrcTestSuite) TestDebugLL() {
	src := `

	`
	res, err := suite.GetLL(src)
	require.NoError(suite.T(), err)
	fmt.Println(res)
}

func (suite *SrcTestSuite) TestPointer() {
	suite.T().Run("Simple Pointer", func(t *testing.T) {
		src := `
	i64 printf(i8 *fmt, ...);

	i64
	main()
	{
		i64 x;
		i64 *p;
		
		x = 4;
		p = &x;
		*p = 0;
		printf("res = %d\n", *p);
		return 0;
	}
	`
		suite.EqualProgramSi(src, "res = 0\n")
	})

	suite.T().Run("Pointer", func(t *testing.T) {
		src := `
	i8* malloc(i64 size);
	i8 free(i8* ptr);
	i8* memset(i8* ptr, i8 val, i64 size);
	i64 printf(i8 *fmt,... );
	i32 strlen(i8* str);

	i64 fib(i64 n) {
		return n + 42;
	}

	i64 main() {
		i8* p = malloc(sizeof(i32) * 42);
		memset(p, (i8)0, sizeof(i32) * 42);
		memset(p, (i8)66, sizeof(i32) * 8);
		
		i8* c = "12345";
		i64 n = 42;
		printf("%s %d,%d", &c[2], sizeof(i32*), strlen(&p[2]));
		free(p);

		return 0;
	}
	`
		suite.EqualProgramSi(src, "345 8,30")
	})
}

func (suite *SrcTestSuite) TestCharCopyPointer() {
	src := `
	i64 printf(i8 *fmt,... );

	i64 main() {
		i8* p = "hello";
		i8* c = p;
		printf("%s", c);
		return 0;
	}
	`
	suite.EqualProgramSi(src, "hello")
}

func (suite *SrcTestSuite) TestStrLen() {
	src := `
	i64 strlen(i8* s);
	i8 printf(i8* format, ...);
	
	i64
	main()
	{
		i8 *p;
		
		p = "hello";
		printf("%d", strlen(p));
		return 0;
	}
	`
	suite.EqualProgramSi(src, "5")
}

func (suite *SrcTestSuite) TestArr() {
	src := `
	i64 printf(i8 *fmt, ...);

	i64 main()
	{
		[2]i64 arr;
		i64 *p;
		
		p = &arr[1];
		*p = 100;
		printf("%d\n",arr[1]);
		return 0;
	}
	`
	suite.EqualProgramSi(src, "100\n")
}

func (suite *SrcTestSuite) TestCond() {
	src := `
i64 printf(i8 *fmt,... );

i64
main()
{
	i64 x = 3;

    if (x == 1)
        printf("(x==1)\n");

	if (x != 1)
		printf("(x != 1)\n");
	else
	   printf("else (x != 1)\n");

	if (x >= 1)
	    printf("(x >= 1)\n");
    else if (x < 1)
        printf("else (x < 1)\n");
    else
        printf("else else (x < 1)\n");

	return 0;
}
	`
	suite.EqualProgramSi(src, "(x != 1)\n(x >= 1)\n")
}

func (suite *SrcTestSuite) TestCalc() {
	src := `
i64 printf(i8 *fmt,... );
i64 main()
{
	i64 x;
	
	x = 1;
	x = x * 10;
	x = x / 2;
	x = x % 3;
	printf("%d", x);
	return 0;
}

`
	suite.EqualProgramSi(src, "2")
}

func (suite *SrcTestSuite) TestStruct() {
	src := `
	type struct { i64 a, i64 b, } my0;

	i64 printf(i8 *fmt,... );

	i64 main() {
		struct { i64 c, i64 d, } m1;
		struct { i64 f, i64 g, } m2;
		
		m1.c = 10;
		m1.d = 20;

		m2.f = 30;
		m2.g = 40;

		m1 = m2;
		
		printf("%d,%d,%d", m1.c, m1.d, sizeof(my0));

		return 0;
	}
	`
	suite.EqualProgramSi(src, "30,40,16")
}

func (suite *SrcTestSuite) TestNull() {
	src := `
	type struct {
		i64 data,
		Node *next,
	} Node;
	
	i64 printf(i8 *fmt,... );
	i8* malloc(i64 size);

	Node* newNode(i64 data) {
		Node* node = (Node*)malloc(sizeof(Node));
		node->data = data;
		node->next = (Node*)NULL;
		return node;
	}

	i64 printList(Node* node) {
		while (node != (Node*)NULL) {
			printf("%d ", node->data);
			node = node->next;
		}
		printf("\n");
		return 0;
	}

	i64 main() {
		Node* head = newNode(1);
		head->next = newNode(2);
		head->next->next = newNode(3);
		head->next->next->next = newNode(4);

		printf("list: ");
		printList(head);
	
		return 0;
	}
	`
	suite.EqualProgramSi(src, "list: 1 2 3 4 \n")
}

func (suite *SrcTestSuite) TestAvoidLeftRecursion() {
	suite.EqualExprSi(`i64 x = 2 * 3 * 4;`, `"%d", x`, "24")
	suite.EqualExprSi(`i64 x = 2;`, `"%d", ++++x`, "4")
	suite.EqualExprSi(`i64 x = 10;`, `"%d,%d", x--, x`, "10,9")
	suite.EqualExprC(`int x = 10; int y = x--;`, `"%d,%d", y, x`, "10,9")
	suite.EqualExprSi(`i64 x = 10;`, `"%d,%d", x----, x`, "9,8")
	suite.ErrorClangExprC(`int x = 10; int b = x----;`, "expression is not assignable")

	// check if expressions are evaluated from left to right
	suite.EqualExprSi(`i64 x = 3 - 1 - 2;`, `"%d", x`, "0")
	suite.EqualExprC(`int x = 3 - 1 - 2;`, `"%d", x`, "0")
	suite.EqualExprSi(`i64 x = 20 / 5 / 2;`, `"%d", x`, "2")
	suite.EqualExprC(`int x = 20 / 5 / 2;`, `"%d", x`, "2")

	suite.EqualExprSi(`i64 x = 3 -3;`, `"%d", x`, "0")
	suite.EqualExprSi(`i64 x = 3 + -3;`, `"%d", x`, "0")
	suite.EqualExprSi(`i64 x = 3 - (-3);`, `"%d", x`, "6")

	suite.EqualExprC(`int x = 3 -3;`, `"%d", x`, "0")
	suite.EqualExprC(`int x = 3 - -3;`, `"%d", x`, "6")
}
