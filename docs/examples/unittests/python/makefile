.PHONY: clean prepare

DOCKER_CALL = docker run --rm -it --net="none"
DOCKER_IMAGE = patwie/test_python3_submission:v3
TASKS = 0.1

clean:
	rm */*.pyc || true
	rm */*.zip || true
	rm *.zip || true
	rm */*.pdf || true

prepare: clean
	cd tasks && pdflatex sheet.tex
	zip -r solution.zip                                                     \
		solution*/*                                                           \
		tasks/*.pdf

	zip -r student_template.zip                                             \
		student_template_*                                                    \
		tasks/*.pdf

	for kind in public private ; do                                         \
	for task in ${TASKS} ; do                                               \
		cd unittest_$${kind}_$${task} &&                                      \
			zip -r unittest_$${kind}_$${task}.zip ./ &&                         \
			cd .. &&                                                            \
			mv unittest_$${kind}_$${task}/unittest_$${kind}_$${task}.zip .  ;   \
	done ;                                                                  \
	done
	for task in ${TASKS} ; do                                               \
		cd solution_$${task} &&                                               \
			zip -r solution_$${task}.zip ./ &&                                  \
			cd .. &&                                                            \
			mv solution_$${task}/solution_$${task}.zip .  ;                     \
		cd student_template_$${task} &&                                       \
			zip -r student_template_$${task}.zip ./ &&                          \
			cd .. &&                                                            \
			mv student_template_$${task}/student_template_$${task}.zip .  ;     \
	done



simulate-template:
	for kind in public private ; do                                         \
	for task in ${TASKS} ; do                                               \
		echo "RUN $${kind} tests $${task}" ;                                  \
		cp student_template_$${task}.zip /tmp/submission.zip ;                \
		cp unittest_$${kind}_$${task}.zip /tmp/unittest.zip ;                 \
		$(DOCKER_CALL)                                                        \
		  -v /tmp/submission.zip:/data/submission.zip:ro                      \
		  -v /tmp/unittest.zip:/data/unittest.zip:ro                          \
		  $(DOCKER_IMAGE);                                                    \
	done;                                                                   \
	done
	rm /tmp/submission.zip
	rm /tmp/unittest.zip

simulate-solution:
	for kind in public private ; do                                         \
	for task in ${TASKS} ; do                                               \
		echo "RUN $${kind} tests $${task}" ;                                  \
		cp solution_$${task}.zip /tmp/submission.zip ;                        \
		cp unittest_$${kind}_$${task}.zip /tmp/unittest.zip ;                 \
		$(DOCKER_CALL)                                                        \
		  -v /tmp/submission.zip:/data/submission.zip:ro                      \
		  -v /tmp/unittest.zip:/data/unittest.zip:ro                          \
		  $(DOCKER_IMAGE);                                                    \
	done;                                                                   \
	done
	rm /tmp/submission.zip
	rm /tmp/unittest.zip