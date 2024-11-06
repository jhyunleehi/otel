# Makefile

## make

### make rule

make는 Target, Depend, Command, Macro 로 구성되어 있습니다.

```
<Target>: <Depend> ?... [[;] <Command>] 
<탭문자><Command> 
```

- Target은 생성하고자 하는 목적물을 지칭
- Depend 는 Target을 만들기 위해서 필요한 요소를 기술
- Command 는 일반 Shell 명령
- Command는 Depend 의 파일생성시간(또는 변경된 시간)을 Target과 비교하여 Target 보다 Depend의 파일이 시간이 보다 최근인 경우로 판단될때에만 실행됩니다
- 주의할것은 Command 는 반드시 앞에 <TAB>문자가 와야 합니다

```makefile
<Makefile>
test: test.o 
        ld -lc -m elf_i386 -dynamic-linker /lib/ld-linux.so.2 -o test /usr/lib/crt1.o /usr/lib/crti.o /usr/lib/crtn.o test.o 
test.o: test.c 
        cc -O2 -Wall -Werror -fomit-frame-pointer -c -o test.o test.c 
```

### Macro

매크로는 다음과 같이 "=" 문자의 왼편에는 Macro의 대표이름(Label)을 기술하고 오른편에는 그 내용을 적습니다. 이때 "=" 문자에 인접한 양쪽의 공백(Space)문자는 무시됩니다.

```makefile
CC = cc 
LD = ld 
CFLAGS = -O2 -Wall -Werror -fomit-frame-pointer -c 
LDFLAGS = -lc -m elf_i386 -dynamic-linker /lib/ld-linux.so.2 
STARTUP = /usr/lib/crt1.o /usr/lib/crti.o /usr/lib/crtn.o 
test: test.o 
        $(LD) $(LDFLAGS) -o test $(STARTUP) test.o 
test.o: test.c 
        $(CC) $(CFLAGS) -o test.o test.c 
```

### 확장규칙 

우리는 보통 C source를 목적파일로 컴파일합니다. 이것은 확장자가 통상 ".c"에서 ".o"를 만들어 내는 규칙이 생성될법 합니다. 

- "$@" 또는 "$(@)"는 바로 Target 을 말합니다. 
- "$<"는 열거된 Depend중에 가장 왼쪽에 기술된 1개의 Depend
- "$^"는 Depend 전체를 의미
- "$?" 로 있는데 이것은 Target과 Depend의 변경날짜를 비교하여 Depend의 변경날짜중에 최근에 변경된것만 선택하는 매크로입니다. "$?"는 주로 라이브러리의 생성 및 관리시에 사용
- 확장자 ".c"를 가진 파일을 확장자 ".o"를 가진 파일로 생성하는 공통적인 확장자 규칙을 예로 작성한 것입니다.

```makefile
CC = cc 
LD = ld 
CFLAGS = -O2 -Wall -Werror -fomit-frame-pointer -c 
LDFLAGS = -lc -m elf_i386 -dynamic-linker /lib/ld-linux.so.2 
STARTUP = /usr/lib/crt1.o /usr/lib/crti.o /usr/lib/crtn.o 
test: test.o 
        $(LD) $(LDFLAGS) -o $@ $(STARTUP) $^ 
.c.o: 
        $(CC) $(CFLAGS) -o $@ $< 
```



#### .PHONY 가짜 target

make clean은 가짜 target이라는 것을 명확하게 정의해 줘야 한다.  혹시 clean target이 있을 수도 있기 때문에...

````makefile
CC = cc 
LD = ld 
RM = rm -f 
CFLAGS = -O2 -Wall -Werror -fomit-frame-pointer -c 
LDFLAGS = -lc -m elf_i386 -dynamic-linker /lib/ld-linux.so.2 
STARTUP = /usr/lib/crt1.o /usr/lib/crti.o /usr/lib/crtn.o 

.PHONY: all clean 

all: test 

clean: 
        $(RM) test.o test 

test: test.o 
        $(LD) $(LDFLAGS) -o $@ $(STARTUP) $^ 

.c.o: 
        $(CC) $(CFLAGS) -o $@ $< 
````



## make helloworld

### test file 

```c
<hello.c>
#include <stdio.h> 
void HelloWorld(void) { 
    fprintf(stdout, "Hello world.\n"); 
} 

<test.c>
#include <stdio.h> 
#include "hello.h" 
int main(void) { 
    HelloWorld(); 
    return(0); 
} 

<hello.h>
extern void HelloWorld(void); 
```

#### Makefile

- make -p option 

```makefile
<Makefile>
CC = cc 
LD = ld 
RM = rm -f 
CFLAGS = -O2 -Wall -Werror -fomit-frame-pointer -v -c 
LDFLAGS = -lc -m elf_i386 -dynamic-linker /lib/ld-linux.so.2 
STARTUP = /usr/lib/crt1.o /usr/lib/crti.o /usr/lib/crtn.o 

BUILD = test 
OBJS = test.o hello.o 

.PHONY: all clean 

all: $(BUILD) 
clean: ; $(RM) *.o $(BUILD) 
test: $(OBJS) ; $(LD) $(LDFLAGS) -o $@ $(STARTUP) $^ 

# 의존관계 성립 
hello.o: $($@:.o=.c) $($@:.o=.h) Makefile 
test.o: $($@:.o=.c) hello.h Makefile 

# 확장자 규칙 (컴파일 공통 규칙) 
.c.o: ; $(CC) $(CFLAGS) -o $@ $< 
```
####  치환
*  위에서 "$($@:.o=.c)" 라는 이상한 문자열
* ` "$(<문자열>:<우측으로부터 매칭될 문자열>=<치환될 문자열>)" `
*  "$@" 부분은 자신의 Target인 "hello.o" 또는 "test.o"를 말합니다. 그리고 거기서 우측으로부터 ".o"가 발견되면 ".c"로 치환하라는 뜻입니다. 
*  "$(hello.o:.o=.c)" 또는 "$(test.o:.o=.c)"로 확장되고 여기서 다시 각각 "hello.c" 와 "test.c"로 치환되어 결국 해당 소스를 지칭하게 되는 셈입니다.

- Command 부분이 <TAB>이 쓰이지 않고 한줄에 ";"(세미콜론)으로 분리되어서 해당 라인에 직접 Command 가 쓰이는 것을 확인하실수 있을겁니다. 무지 거대한 "Makefile"을 간략히 보이게 하기 위해서 이렇게도 사용할수 있다는 것을 예로 보인것입니다. 의존관계를 성립하는 부분은 Command 가 없는것을 볼수 있는데 이것은 비슷한 다른 Target에서 Command 가 결합되어 수행될수 있고 여기서는 ".c.o: ; ..." 부분의 Command 가 결합됩니다. 여기서 의존관계를 최대한 자세하게 기술하였는데 만약 "hello.h" 가 변경된다면 "hello.o"와 "test.o"는 다시 빌드될것입니다. 또한 "Makefile" 도 수정되면 다시 빌드될것이라는 것이 예상됩니다. 이처럼 의존관계를 따로 기술하는 이유는 차후에 여러분들이 사용하시다보면 이유를 알게 될겁니다. 의존관계라는게 서로 굉장히 유기적으로 걸리는 경우가 많기 때문에 보다 보기 편하게 하는 이유도 있고 차후에 의존관계가 변경되었을때 쉽게 찾아서 변경을 할수 있도록 하는것도 한가지 이유입니다.

#### makefile

```makefile
CC = cc 
LD = ld 
RM = rm -f 
CFLAGS = -O2 -Wall -Werror -fomit-frame-pointer -v -c 
LDFLAGS = -lc -m elf_i386 -dynamic-linker /lib/ld-linux.so.2 
STARTUP = /usr/lib/crt1.o /usr/lib/crti.o /usr/lib/crtn.o 

BUILD = test 
OBJS = test.o hello.o 

.PHONY: all clean 

all: $(BUILD) 
clean: ; $(RM) *.o $(BUILD) 
test: $(OBJS) ; $(LD) $(LDFLAGS) -o $@ $(STARTUP) $^ 

# 의존관계 성립 
$(OBJS): $($@:.o=.c) hello.h Makefile 
# test.o hello.o: $($@:.o=.c) hello.h Makefile 

# 확장자 규칙 (컴파일 공통 규칙) 
.c.o: ; $(CC) $(CFLAGS) -o $@ $< 
```

####  매크로 치환 
매크로를 지정하고, 그것을 이용하는 것을 이미 알고 있다. 그런데, 필요에 의해 이미 매크로의 내용을 조그만 바꾸어야 할 때가 있다. 매크로 내용의 일부만 바꾸기 위해서는 $(MACRO_NAME:OLD=NEW)과 같은 형식을 이용하면 된다.
```
MY_NAME = Michael Jackson
YOUR_NAME = $(NAME:Jack=Jook)
```
위의 예제에서는 Jack이란 부분이 Jook으로 바뀌게 된다. 즉 YOUR_NAME 이란 매크로의 값은 Michael Jookson 이 된다. 아래의 예제를 하나 더 보기로 한다.
```
OBJS = main.o read.o write.o
SRCS = $(OBJS:.o=.c)
```
위의 예제에서는 OBJS에서 .c가 .o로 바뀌게 된다. 즉 아래와 같다.
```
SRCS = main.c read.c write.c
```
위의 예제는 실제로 사용하면 아주 편할 때가 많다. 가령 .o 파일 100개에 .c 파일이 각각 있을 때 이들을 다 적으려면 무척이나 짜증나는 일이 될 것이다.
