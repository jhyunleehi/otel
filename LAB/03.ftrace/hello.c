#include <stdio.h>
char filename[] = "output.txt";
char message[] = "hello, world!\n";
FILE * file;
void rcall(int n){
    if (n<0) return;
    fputs( message, file);    
    fprintf(file,"[%d]:%s", n, message);
    rcall(n-1);
}
int main()
{
    file = fopen(filename, "w");
    if (file == NULL) return -1;    
    rcall(10);
    fclose(file);
    return 0;
}