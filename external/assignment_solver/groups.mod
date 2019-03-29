set student;
set group;

var assign{i in student, j in group} binary;
param pref{i in student, j in group};

maximize totalPref:
     sum{i in student, j in group} pref[i,j]*assign[i,j];

subject to exactly_one_group {i in student}:
      sum {j in group} assign[i,j] =1;

subject to min3{j in group}:
     sum{i in student} assign[i,j]>=3;

subject to max4{j in group}:
     sum{i in student} assign[i,j]<=4;

end;